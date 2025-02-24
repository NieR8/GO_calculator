package agent

import (
	"fmt"
	"github.com/NieR8/myProject/internal/orchestrator"
	"time"
)

//type Orchestrator struct {
//	tasks   []Task
//	agents  []*Agent
//	results map[string]float64
//	mutex   sync.Mutex
//	wg      sync.WaitGroup
//}

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
	orch     *orchestrator.Orchestrator
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
