package parser

import (
	"github.com/NieR8/myProject/models"
	"testing"
)

func TestInfixToRPN(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2+3", "2 3 +"},
		{"(5+2)+4/5", "5 2 + 4 5 / +"},
		{"2++3", "2 + 3 +"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := InfixToRPN(tt.input)
			if err != nil {
				t.Errorf("InfixToRPN(%q) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("InfixToRPN(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildTasks(t *testing.T) {
	tests := []struct {
		exprID   string
		root     *models.Node
		expected []models.Task
		wantErr  bool
	}{
		{
			"expr-1",
			&models.Node{Value: "+", Left: &models.Node{Value: "2"}, Right: &models.Node{Value: "3"}},
			[]models.Task{{ID: "task-expr-1-0", Arg1: "2", Arg2: "3", Operation: "+", Completed: false}},
			false,
		},
		{
			"expr-1",
			&models.Node{Value: "/", Left: &models.Node{Value: "4"}, Right: &models.Node{Value: "0"}},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.exprID, func(t *testing.T) {
			tasks, err := BuildTasks(tt.exprID, tt.root)
			if tt.wantErr {
				if err == nil {
					t.Errorf("BuildTasks(%q) expected error, got %v", tt.exprID, err)
				}
			} else {
				if err != nil {
					t.Errorf("BuildTasks(%q) unexpected error: %v", tt.exprID, err)
				}
				if len(tasks) != len(tt.expected) {
					t.Errorf("BuildTasks(%q) returned %d tasks, want %d", tt.exprID, len(tasks), len(tt.expected))
				}
				for i, task := range tasks {
					if task.ID != tt.expected[i].ID || task.Arg1 != tt.expected[i].Arg1 || task.Arg2 != tt.expected[i].Arg2 || task.Operation != tt.expected[i].Operation {
						t.Errorf("BuildTasks(%q) task %d = %+v, want %+v", tt.exprID, i, task, tt.expected[i])
					}
				}
			}
		})
	}
}
