package calculation

import (
	"strconv"
	"strings"
)

// Calc вычисляет выражение и возвращает результат
func Calc(expression string) (float64, error) {
	tokens, err := tokenize(expression)
	if len(tokens) < 2 {
		return 0, ErrNotEnoughArgs
	}
	postfix, err := infixToPostfix(tokens)
	result, err := evalPostfix(postfix)
	return result, err
}

// tokenize разбивает строку на токены (числа и операторы)
func tokenize(expression string) ([]string, error) {
	var tokens []string
	var current strings.Builder

	for i, char := range expression {
		if char >= '0' && char <= '9' || char == '.' {
			current.WriteRune(char)
		} else {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			if char == ' ' {
				continue
			}
			if char == '+' || char == '-' || char == '*' || char == '/' || char == '(' || char == ')' {
				if char == '-' && (i == 0 || (i > 0 && (expression[i-1] == '(' || strings.ContainsAny(string(expression[i-1]), "+-*/")))) {
					current.WriteRune(char) // добавляем знак минус к числу
				} else {
					tokens = append(tokens, string(char))
				}

			} else {
				return nil, ErrInvalidSymbol // ошибка: недопустимый символ
			}
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens, nil
}

// infixToPostfix преобразует выражение из инфиксной записи в постфиксную
func infixToPostfix(tokens []string) ([]string, error) {
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
			output = append(output, token) // если токен - число
		} else if token == "(" {
			stack = append(stack, token)
		} else if token == ")" {
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return nil, ErrOpenedParentheses // ошибка: отсутствующая открывающая скобка
			}
			stack = stack[:len(stack)-1] // убираем открывающую скобку
		} else {
			for len(stack) > 0 && precedence[stack[len(stack)-1]] >= precedence[token] {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token) // добавляем оператор
		}
	}

	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, ErrClosedParentheses // ошибка: отсутствующая закрывающая скобка
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return output, nil
}

// evalPostfix вычисляет значение постфиксного выражения
func evalPostfix(tokens []string) (float64, error) {
	stack := []float64{}

	for _, token := range tokens {
		if num, err := strconv.ParseFloat(token, 64); err == nil {
			stack = append(stack, num) // если токен - число
		} else {
			if len(stack) < 2 {
				return 0, ErrNotEnoughArgs
			}
			b, a := stack[len(stack)-1], stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			switch token {
			case "+":
				stack = append(stack, a+b)
			case "-":
				stack = append(stack, a-b)
			case "*":
				stack = append(stack, a*b)
			case "/":
				if b == 0 {
					return 0, ErrDivisionByZero
				}
				stack = append(stack, a/b)
			default:
				return 0, ErrInvalidOperation
			}
		}
	}

	if len(stack) != 1 {
		return 0, ErrInvalidExpression
	}
	return stack[0], nil
}
