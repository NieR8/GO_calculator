package models

// Node представляет узел дерева операций
type Node struct {
	Value string `json:"value"`
	Left  *Node  `json:"left,omitempty"`
	Right *Node  `json:"right,omitempty"`
}

// Task представляет задачу для вычисления
type Task struct {
	ID        string  `json:"id"`
	Arg1      string  `json:"arg1"`
	Arg2      string  `json:"arg2"`
	Operation string  `json:"operation"` // (+ - / *)
	Result    float64 `json:"result,omitempty"`
	Completed bool    `json:"completed"`
}

// Result представляет результат выполнения задачи
type Result struct {
	TaskID string  `json:"task_id"`
	Value  float64 `json:"value"`
	Error  string  `json:"error,omitempty"`
}

// Expression представляет арифметическое выражение
type Expression struct {
	Name   string  `json:"name"`
	Status int     `json:"status"` // 0: посчиталось, 1: считается, 2: ожидает вычисления, 3: невалидно
	Id     int     `json:"id"`
	Result float64 `json:"result"`
	Node   *Node   `json:"node,omitempty"`
}
