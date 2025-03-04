package parser

import (
	"fmt"
	"github.com/NieR8/myProject/models"
	"log"
	"strconv"
	"strings"
)

func CheckExpression(str []string) bool {
	for i := 0; i < len(str)-1; i++ {
		switch {
		case str[i] == "(" && str[i+1] == ")":
			return false // Проверка на пустые скобки
		}
	}
	return true
}

// Разбивает строку на токены (числа и операторы)
func tokenize(expression string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	expression = strings.ReplaceAll(expression, " ", "") // Удаляем пробелы

	for i, char := range expression {
		if char >= '0' && char <= '9' || char == '.' {
			current.WriteRune(char)
		} else {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				if current.Len() > 1 && current.String()[0] == '0' && current.String()[1] != '.' {
					return nil, ErrInvalidSymbol // Проверка на ведущий ноль
				}
				current.Reset()
			}
			if char == '+' || char == '-' || char == '*' || char == '/' || char == '(' || char == ')' {
				if char == '-' && (i == 0 || (i > 0 && expression[i-1] == '(')) {
					current.WriteRune(char) // Унарный минус
				} else {
					tokens = append(tokens, string(char))
				}
			} else {
				return nil, ErrInvalidSymbol // Недопустимый символ
			}
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	for _, v := range tokens {
		if len(v) > 1 && v[0] == '0' && v[1] != '.' || len(v) > 1 && v[0] == '-' && v[1] == '0' {
			return nil, ErrInvalidSymbol // Проверка на ведущий ноль после минуса
		}
	}

	if !CheckExpression(tokens) {
		return nil, ErrInvalidExpression
	}

	return tokens, nil
}

// Преобразует выражение из инфиксной записи в постфиксную (RPN)
func InfixToRPN(expression string) (string, error) {
	tokens, err := tokenize(expression)
	if err != nil {
		return "", err
	}

	var output []string
	var stack []string

	precedence := map[string]int{
		"+": 1,
		"-": 1,
		"*": 2,
		"/": 2,
	}

	for _, token := range tokens {
		if _, err := strconv.ParseFloat(token, 64); err == nil {
			output = append(output, token)
		} else if token == "(" {
			stack = append(stack, token)
		} else if token == ")" {
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return "", ErrInvalidExpression // Нет открывающей скобки
			}
			stack = stack[:len(stack)-1]
		} else {
			for len(stack) > 0 && precedence[stack[len(stack)-1]] >= precedence[token] {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		}
	}

	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return "", ErrInvalidExpression // Нет закрывающей скобки
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return strings.Join(output, " "), nil
}

// Парсит RPN выражение и строит дерево операций
func ParseRPN(rpn string) (*models.Node, error) {
	if rpn == "" {
		return nil, ErrEmptyExpression
	}

	tokens := strings.Fields(rpn)
	stack := make([]*models.Node, 0)

	for _, token := range tokens {
		if IsOperator(token) {
			if len(stack) < 2 {
				return nil, ErrInvalidRpn
			}
			right := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			left := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			node := &models.Node{Value: token, Left: left, Right: right}
			stack = append(stack, node)
		} else {
			num, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return nil, ErrInvalidSymbol
			}
			node := &models.Node{Value: fmt.Sprintf("%f", num)}
			stack = append(stack, node)
		}
	}

	if len(stack) != 1 {
		return nil, fmt.Errorf("invalid RPN expression")
	}

	return stack[0], nil
}

func IsOperator(token string) bool {
	return token == "+" || token == "-" || token == "*" || token == "/"
}

// Строит список задач на основе дерева
func BuildTasks(exprID string, root *models.Node) ([]models.Task, error) {
	if root == nil {
		return nil, ErrEmptyExpression
	}

	var tasks []models.Task
	var taskCounter int

	var buildTask func(node *models.Node) (string, error)
	buildTask = func(node *models.Node) (string, error) {
		if node == nil {
			return "", nil
		}

		if !IsOperator(node.Value) {
			return node.Value, nil // Число
		}

		if node.Value == "/" {
			if rightNum, err := strconv.ParseFloat(node.Right.Value, 64); err == nil && rightNum == 0 {
				return "", ErrDivisionByZero
			}
		}

		leftArg, err := buildTask(node.Left)
		if err != nil {
			return "", err
		}
		rightArg, err := buildTask(node.Right)
		if err != nil {
			return "", err
		}

		taskID := fmt.Sprintf("task-%s-%d", exprID, taskCounter)
		taskCounter++

		task := models.Task{
			ID:        taskID,
			Arg1:      leftArg,
			Arg2:      rightArg,
			Operation: node.Value,
			Completed: false,
		}
		tasks = append(tasks, task)
		return taskID, nil
	}

	_, err := buildTask(root)
	if err != nil {
		return nil, err // Возвращаем ошибку
	}

	for i, j := 0, len(tasks)-1; i < j; i, j = i+1, j-1 {
		tasks[i], tasks[j] = tasks[j], tasks[i] // Разворачиваем задачи, чтобы дерево считалось снизу вверх
	}
	log.Printf("Сформированы задачи для %s: %+v", exprID, tasks)
	return tasks, nil
}
