package store

import (
	"github.com/NieR8/myProject/models"
	"testing"
)

func TestAddAndGetExpression(t *testing.T) {
	store := NewStore()
	expr := models.Expression{Name: "2+3", Status: 1, Id: 1}
	store.AddExpression(expr)

	got, exists := store.GetExpression(1)
	if !exists {
		t.Errorf("GetExpression(1) expected to exist")
	}
	if got.Name != expr.Name || got.Status != expr.Status {
		t.Errorf("GetExpression(1) = %+v, want %+v", got, expr)
	}
}

func TestUpdateTask(t *testing.T) {
	store := NewStore()
	expr := models.Expression{
		Name:   "2+3",
		Status: 1,
		Id:     1,
		Node:   &models.Node{Value: "+", Left: &models.Node{Value: "2"}, Right: &models.Node{Value: "3"}},
	}
	store.AddExpression(expr)
	task := models.Task{ID: "task-expr-1-0", Arg1: "2", Arg2: "3", Operation: "+"}
	store.Tasks[task.ID] = task

	result := models.Result{TaskID: "task-expr-1-0", Value: 5}
	success := store.UpdateTask(result)
	if !success {
		t.Errorf("UpdateTask failed")
	}

	updatedTask, exists := store.Tasks["task-expr-1-0"]
	if !exists || updatedTask.Result != 5 || !updatedTask.Completed {
		t.Errorf("Task not updated correctly: %+v", updatedTask)
	}

	updatedExpr, _ := store.GetExpression(1)
	if updatedExpr.Status != 0 || updatedExpr.Result != 5 {
		t.Errorf("Expression not completed: %+v", updatedExpr)
	}
}
