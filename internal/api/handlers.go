package api

import (
	"encoding/json"
	"github.com/NieR8/myProject/internal/store"
	"github.com/NieR8/myProject/models"
	"log"
	"net/http"
	"strings"
)

func HandleTask(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path == "/internal/task" {
				handleGetTask(w, r, st)
			} else {
				http.Error(w, "Not found", http.StatusNotFound)
			}
		case http.MethodPost:
			handlePostTask(w, r, st)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func HandleTaskResult(st *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handleGetTaskResult(w, r, st)
	}
}

func handleGetTask(w http.ResponseWriter, r *http.Request, st *store.Store) {
	task, exists := st.GetPendingTask()
	if !exists {
		http.Error(w, "No task available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Task models.Task `json:"task"`
	}{Task: task})
}

func handlePostTask(w http.ResponseWriter, r *http.Request, st *store.Store) {
	var result models.Result
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		log.Printf("Ошибка декодирования результата: %v", err)
		http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
		return
	}

	log.Printf("Получен результат для задачи %s: %f", result.TaskID, result.Value)
	if result.TaskID == "" {
		log.Println("Отсутствует TaskID в результате")
		http.Error(w, "Missing task ID", http.StatusUnprocessableEntity)
		return
	}

	if !st.UpdateTask(result) {
		log.Printf("Задача %s не найдена при обновлении", result.TaskID)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	log.Printf("Результат задачи %s успешно принят: %f", result.TaskID, result.Value)
	w.WriteHeader(http.StatusOK)
}

func handleGetTaskResult(w http.ResponseWriter, r *http.Request, st *store.Store) {
	taskID := strings.TrimPrefix(r.URL.Path, "/internal/task/result/")
	if taskID == "" {
		log.Println("Отсутствует taskID в запросе результата")
		http.Error(w, "Missing task ID", http.StatusBadRequest)
		return
	}

	st.Mu.Lock()
	defer st.Mu.Unlock()

	task, exists := st.Tasks[taskID]
	if !exists {
		log.Printf("Задача %s не найдена в Tasks", taskID)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if !task.Completed {
		log.Printf("Результат задачи %s ещё не готов: %+v", taskID, task)
		http.Error(w, "Task result not available", http.StatusNotFound)
		return
	}

	log.Printf("Возвращён результат задачи %s: %f", taskID, task.Result)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Result float64 `json:"result"`
	}{Result: task.Result})
}
