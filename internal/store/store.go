package store

import (
	"fmt"
	"github.com/NieR8/myProject/models"
	"github.com/NieR8/myProject/pkg/parser"
	"log"
	"strconv"
	"strings"
	"sync"
)

type Store struct {
	Mu           sync.Mutex
	Expressions  map[int]models.Expression
	Tasks        map[string]models.Task
	PendingTasks chan models.Task
}

func NewStore() *Store {
	return &Store{
		Expressions:  make(map[int]models.Expression),
		Tasks:        make(map[string]models.Task),
		PendingTasks: make(chan models.Task, 100),
	}
}

func (s *Store) AddExpression(expr models.Expression) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Expressions[expr.Id] = expr
	log.Printf("Добавлено выражение %d: %+v", expr.Id, expr)
}

func (s *Store) GetExpression(id int) (models.Expression, bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	expr, exists := s.Expressions[id]
	log.Printf("Запрошено выражение %d: найдено=%v, %+v", id, exists, expr)
	return expr, exists
}

func (s *Store) GetAllExpressions() []models.Expression {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	var expressions []models.Expression
	for _, expr := range s.Expressions {
		expressions = append(expressions, expr)
	}
	log.Printf("Возвращено %d выражений", len(expressions))
	return expressions
}

func (s *Store) AddTask(task models.Task) {
	s.Mu.Lock()
	s.Tasks[task.ID] = task
	log.Printf("Задача %s добавлена в Tasks: %+v, всего задач: %d", task.ID, task, len(s.Tasks))
	s.PendingTasks <- task
	s.Mu.Unlock()
}

func (s *Store) UpdateTask(result models.Result) bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	task, exists := s.Tasks[result.TaskID]
	if !exists {
		log.Printf("Ошибка: задача %s не найдена в Tasks: %+v", result.TaskID, s.Tasks)
		return false
	}

	log.Printf("Обновление задачи %s: старое значение %+v, новый результат %f", result.TaskID, task, result.Value)
	task.Result = result.Value
	task.Completed = true // Устанавливаем флаг завершения
	s.Tasks[result.TaskID] = task
	log.Printf("Задача %s обновлена: %+v", result.TaskID, task)

	parts := strings.Split(result.TaskID, "-")
	if len(parts) < 3 {
		log.Printf("Ошибка: неверный формат TaskID %s", result.TaskID)
		return false
	}
	exprIDStr := parts[2]
	id, err := strconv.Atoi(exprIDStr)
	if err != nil {
		log.Printf("Ошибка разбора exprID из %s: %v", result.TaskID, err)
		return false
	}

	expr, exists := s.Expressions[id]
	if !exists {
		log.Printf("Выражение %d не найдено для задачи %s", id, result.TaskID)
		return false
	}

	allCompleted := true
	for _, t := range s.Tasks {
		if strings.Contains(t.ID, fmt.Sprintf("expr-%d", id)) && !t.Completed {
			log.Printf("Задача %s для выражения %d ещё не завершена: %+v", t.ID, id, t)
			allCompleted = false
			break
		}
	}

	if allCompleted {
		log.Printf("Все задачи для выражения %d завершены, пересчитываем результат", id)
		finalResult, err := s.calculateExpression(expr)
		if err != nil {
			expr.Status = 3
			s.Expressions[id] = expr
			log.Printf("Ошибка при вычислении выражения %d: %v", id, err)
			return true
		}
		expr.Result = finalResult
		expr.Status = 0
		s.Expressions[id] = expr
		log.Printf("Выражение %d завершено: %+v", id, expr)
	}

	return true
}

func (s *Store) calculateExpression(expr models.Expression) (float64, error) {
	return s.evaluateNode(expr.Node)
}

func (s *Store) evaluateNode(node *models.Node) (float64, error) {
	if node == nil {
		return 0, fmt.Errorf("nil node")
	}

	if !parser.IsOperator(node.Value) {
		return strconv.ParseFloat(node.Value, 64)
	}

	leftVal, err := s.evaluateNode(node.Left)
	if err != nil {
		return 0, err
	}

	rightVal, err := s.evaluateNode(node.Right)
	if err != nil {
		return 0, err
	}

	switch node.Value {
	case "+":
		return leftVal + rightVal, nil
	case "-":
		return leftVal - rightVal, nil
	case "*":
		return leftVal * rightVal, nil
	case "/":
		if rightVal == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return leftVal / rightVal, nil
	default:
		return 0, fmt.Errorf("unsupported operation: %s", node.Value)
	}
}

func (s *Store) GetPendingTask() (models.Task, bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	for i := 0; i < cap(s.PendingTasks); i++ {
		select {
		case task := <-s.PendingTasks:
			if s.isTaskReady(task) {
				log.Printf("Задача %s готова и выдана: %+v", task.ID, task)
				return task, true
			}
			log.Printf("Задача %s не готова, возвращаем в очередь: %+v", task.ID, task)
			s.PendingTasks <- task
		default:
			//log.Println("Очередь задач пуста")
			return models.Task{}, false
		}
	}
	log.Println("Нет готовых задач в очереди после проверки")
	return models.Task{}, false
}

func (s *Store) isTaskReady(task models.Task) bool {
	if isNumeric(task.Arg1) && isNumeric(task.Arg2) {
		return true
	}
	if !isNumeric(task.Arg1) {
		depTask, exists := s.Tasks[task.Arg1]
		if !exists || !depTask.Completed {
			log.Printf("Зависимость %s для задачи %s не готова: exists=%v, completed=%v", task.Arg1, task.ID, exists, depTask.Completed)
			return false
		}
	}
	if !isNumeric(task.Arg2) {
		depTask, exists := s.Tasks[task.Arg2]
		if !exists || !depTask.Completed {
			log.Printf("Зависимость %s для задачи %s не готова: exists=%v, completed=%v", task.Arg2, task.ID, exists, depTask.Completed)
			return false
		}
	}
	return true
}

func isNumeric(arg string) bool {
	_, err := strconv.ParseFloat(arg, 64)
	return err == nil
}
