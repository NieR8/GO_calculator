package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/NieR8/myProject/internal/api"
	"github.com/NieR8/myProject/internal/store"
	"github.com/NieR8/myProject/models"
	"github.com/NieR8/myProject/pkg/parser"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
)

type Orchestrator struct {
	Addr        string
	Server      *http.Server
	Store       *store.Store
	taskCounter uint64
}

func NewOrchestrator(addr string) *Orchestrator {
	st := store.NewStore()
	return &Orchestrator{
		Addr:  addr,
		Store: st,
		Server: &http.Server{
			Addr:    addr,
			Handler: nil,
		},
	}
}

func (o *Orchestrator) Run(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/calculate", o.handleCalculate)
	mux.HandleFunc("/api/v1/expressions", o.handleGetExpressions)
	mux.HandleFunc("/api/v1/expressions/", o.handleGetExpressionByID)
	mux.HandleFunc("/internal/task", api.HandleTask(o.Store))
	mux.HandleFunc("/internal/task/result/", api.HandleTaskResult(o.Store))
	mux.HandleFunc("/api/v1/pending-tasks", o.handleGetPendingTasks) // эндпоинт для мониторинга еще незавершенных задач

	o.Server.Handler = mux

	go func() {
		if err := o.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Ошибка сервера: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Останавливаем оркестратор")
	return o.Server.Shutdown(context.Background())
}

// Принимает POST-запросы, парсит выражение, создаёт задачи и добавляет их в очередь
func (o *Orchestrator) handleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Expression string `json:"expression"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Expression == "" {
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	id := int(atomic.AddUint64(&o.taskCounter, 1))
	expr := models.Expression{
		Name:   req.Expression,
		Status: 2,
		Id:     id,
	}

	o.Store.AddExpression(expr)
	log.Printf("Выражение %d со статусом 2 добавлено в Store: %+v", id, expr)

	rpn, err := parser.InfixToRPN(req.Expression)
	if err != nil {
		expr.Status = 3
		o.Store.AddExpression(expr)
		http.Error(w, "Invalid expression: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	tree, err := parser.ParseRPN(rpn)
	if err != nil {
		expr.Status = 3
		o.Store.AddExpression(expr)
		http.Error(w, "Failed to parse expression", http.StatusUnprocessableEntity)
		return
	}

	expr.Node = tree
	tasks, err := parser.BuildTasks(fmt.Sprintf("expr-%d", id), tree)
	if err != nil {
		expr.Status = 3
		o.Store.AddExpression(expr)
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if len(tasks) == 0 && tree != nil && !parser.IsOperator(tree.Value) { // Если задач нет и это просто одно число
		result, err := strconv.ParseFloat(tree.Value, 64)
		if err != nil {
			expr.Status = 3
			o.Store.AddExpression(expr)
			http.Error(w, "Invalid number: "+err.Error(), http.StatusUnprocessableEntity)
			return
		}
		expr.Status = 0
		expr.Result = result
		o.Store.AddExpression(expr)
		log.Printf("Выражение %d завершено без задач: %+v", id, expr)
	} else {
		expr.Status = 1
		for i := len(tasks) - 1; i >= 0; i-- {
			o.Store.AddTask(tasks[i])
		}
		o.Store.AddExpression(expr)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		ID int `json:"id"`
	}{ID: id})
}

// Возвращает список всех выражений
func (o *Orchestrator) handleGetExpressions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	expressions := o.Store.GetAllExpressions()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Expressions []models.Expression `json:"expressions"`
	}{Expressions: expressions})
}

// Возвращает конкретное выражение
func (o *Orchestrator) handleGetExpressionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")
	id, err := strconv.Atoi(idStr)
	if err != nil || idStr == "" {
		http.Error(w, "Invalid or missing ID", http.StatusBadRequest)
		return
	}

	expr, exists := o.Store.GetExpression(id)
	if !exists {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Expression models.Expression `json:"expression"`
	}{Expression: expr})
}

func (o *Orchestrator) handleGetPendingTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tasks []models.Task
	o.Store.Mu.Lock()
	//for i := 0; i < cap(o.Store.PendingTasks); i++ {
	//	select {
	//	case task := <-o.Store.PendingTasks:
	//		tasks = append(tasks, task)
	//		o.Store.PendingTasks <- task // Возвращаем задачу в очередь
	//	default:
	//		break
	//	}
	//}
	for _, task := range o.Store.Tasks {
		if !task.Completed { // Показываем только незавершённые задачи
			tasks = append(tasks, task)
		}
	}
	o.Store.Mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(struct {
		Tasks []models.Task `json:"tasks"`
	}{Tasks: tasks})
	if err != nil {
		log.Printf("Ошибка сериализации задач: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
