package orchestrator

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Node представляет узел дерева операций
type Node struct {
	Value string
	Left  *Node
	Right *Node
}

// Task представляет задачу с новой структурой
type Task struct {
	ID            string  `json:"id"`
	Arg1          string  `json:"arg1"`
	Arg2          string  `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime float64 `json:"operation_time"`
	Result        float64 `json:"result,omitempty"`
}

// Agent представляет исполнителя задач
type Agent struct {
	ID       string
	Busy     bool
	TaskChan chan *Task
	orch     *Orchestrator
}

// Orchestrator управляет задачами и агентами
type Orchestrator struct {
	tasks   []Task
	agents  []*Agent
	results map[string]float64
	mutex   sync.Mutex
	wg      sync.WaitGroup
}

// ExpressionRequest представляет входящий JSON-запрос
type ExpressionRequest struct {
	Expression string `json:"expression"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}

// NewOrchestrator создаёт новый оркестратор
func NewOrchestrator(numAgents int) *Orchestrator {
	agents := make([]*Agent, numAgents)
	for i := 0; i < numAgents; i++ {
		agents[i] = &Agent{
			ID:       fmt.Sprintf("agent_%d", i),
			Busy:     false,
			TaskChan: make(chan *Task),
		}
		go agents[i].run()
	}
	return &Orchestrator{
		agents:  agents,
		results: make(map[string]float64),
	}
}

// ParseRPN парсит RPN и строит дерево операций
func ParseRPN(rpn string) (*Node, error) {
	if rpn == "" {
		return nil, fmt.Errorf("пустое выражение")
	}

	tokens := strings.Fields(rpn)
	stack := make([]*Node, 0)

	for _, token := range tokens {
		if isOperator(token) {
			if len(stack) < 2 {
				return nil, fmt.Errorf("некорректное RPN-выражение")
			}
			// Извлекаем правый и левый операнды
			right := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			left := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			// Создаём новый узел с оператором
			node := &Node{Value: token, Left: left, Right: right}
			stack = append(stack, node)
		} else {
			// Если токен — число, создаём лист
			num, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return nil, fmt.Errorf("некорректное число: %s", token)
			}
			node := &Node{Value: fmt.Sprintf("%f", num), Left: nil, Right: nil}
			stack = append(stack, node)
		}
	}

	if len(stack) != 1 {
		return nil, fmt.Errorf("некорректное RPN-выражение")
	}

	return stack[0], nil
}

// isOperator проверяет, является ли токен оператором
func isOperator(token string) bool {
	return token == "+" || token == "-" || token == "*" || token == "/"
}

// generateID генерирует уникальный идентификатор задачи
func generateID() string {
	return fmt.Sprintf("task_%d", rand.Intn(1000))
}

// BuildTasks строит список задач на основе дерева
func (o *Orchestrator) BuildTasks(root *Node, taskCounter *int) error {
	if root == nil {
		return fmt.Errorf("пустое выражение")
	}

	var buildTask func(node *Node) *Task
	buildTask = func(node *Node) *Task {
		if node == nil {
			return nil
		}

		*taskCounter++
		taskID := generateID()
		task := Task{
			ID:            taskID,
			OperationTime: rand.Float64() * 0.1, // Случайное время выполнения (0-0.1 сек)
		}

		if !isOperator(node.Value) {
			// Если операнд (число)
			num, _ := strconv.ParseFloat(node.Value, 64)
			task.Arg1 = node.Value
			task.Arg2 = ""
			task.Result = num
			task.Operation = "value"
			return &task
		}

		// Если оператор, создаём задачи для левого и правого поддеревьев
		task.Operation = node.Value
		leftTask := buildTask(node.Left)
		rightTask := buildTask(node.Right)

		if leftTask != nil {
			task.Arg1 = leftTask.ID
		}
		if rightTask != nil {
			task.Arg2 = rightTask.ID
		}

		return &task
	}

	// Добавляем корневую задачу
	rootTask := buildTask(root)
	o.tasks = append(o.tasks, *rootTask)
	return nil
}

// run запускает агента для обработки задач
func (a *Agent) run() {
	for task := range a.TaskChan {
		a.Busy = true
		time.Sleep(time.Duration(task.OperationTime * float64(time.Second))) // Симулируем выполнение
		result, err := a.executeTask(task)
		if err == nil {
			a.orch.mutex.Lock()
			a.orch.results[task.ID] = result
			a.orch.mutex.Unlock()
		}
		a.Busy = false
		a.orch.wg.Done()
	}
}

// executeTask выполняет задачу
func (a *Agent) executeTask(t *Task) (float64, error) {
	if t.Operation == "value" {
		return t.Result, nil
	}

	var leftVal, rightVal float64

	if t.Arg1 != "" {
		a.orch.mutex.Lock()
		if val, exists := a.orch.results[t.Arg1]; exists {
			leftVal = val
		} else {
			a.orch.mutex.Unlock()
			return 0, fmt.Errorf("результат для задачи %s не найден", t.Arg1)
		}
		a.orch.mutex.Unlock()
	}
	if t.Arg2 != "" {
		a.orch.mutex.Lock()
		if val, exists := a.orch.results[t.Arg2]; exists {
			rightVal = val
		} else {
			a.orch.mutex.Unlock()
			return 0, fmt.Errorf("результат для задачи %s не найден", t.Arg2)
		}
		a.orch.mutex.Unlock()
	}

	switch t.Operation {
	case "+":
		t.Result = leftVal + rightVal
	case "-":
		t.Result = leftVal - rightVal
	case "*":
		t.Result = leftVal * rightVal
	case "/":
		if rightVal == 0 {
			return 0, fmt.Errorf("деление на ноль")
		}
		t.Result = leftVal / rightVal
	default:
		return 0, fmt.Errorf("неизвестная операция: %s", t.Operation)
	}

	return t.Result, nil
}

// ExecuteTasks распределяет задачи между агентами
func (o *Orchestrator) ExecuteTasks() (float64, error) {
	if len(o.tasks) == 0 {
		return 0, fmt.Errorf("нет задач для выполнения")
	}

	o.wg.Add(len(o.tasks))
	for i := range o.tasks {
		// Найти свободного агента
		var agent *Agent
		for {
			for _, a := range o.agents {
				if !a.Busy {
					agent = a
					break
				}
			}
			if agent != nil {
				break
			}
			time.Sleep(10 * time.Millisecond) // Ждём, если все агенты заняты
		}
		agent.TaskChan <- &o.tasks[i]
	}

	o.wg.Wait()
	// Получаем результат корневой задачи
	result, exists := o.results[o.tasks[0].ID]
	if !exists {
		return 0, fmt.Errorf("результат корневой задачи не найден")
	}
	return result, nil
}

// calculateHandler обрабатывает запросы API
func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req ExpressionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Парсим RPN
	tree, err := ParseRPN(req.Expression)
	if err != nil {
		resp := ErrorResponse{Error: "ошибка в выражении"}
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Создаём оркестратор с 3 агентами (можно настроить)
	orch := NewOrchestrator(3)
	var taskCounter int
	if err := orch.BuildTasks(tree, &taskCounter); err != nil {
		resp := ErrorResponse{Error: "не удалось построить задачи"}
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Выполняем задачи
	result, err := orch.ExecuteTasks()
	if err != nil {
		resp := ErrorResponse{Error: err.Error()}
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Возвращаем результат и список задач
	response := struct {
		Result float64 `json:"result"`
		Tasks  []Task  `json:"tasks"`
	}{
		Result: result,
		Tasks:  orch.tasks,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	rand.Seed(time.Now().UnixNano()) // Для генерации случайных ID
	http.HandleFunc("/api/v1/calculate", CalculateHandler)
	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
