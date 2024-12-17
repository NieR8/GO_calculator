package calculation

import (
	"strconv"
	"strings"
)

func CheckExpression(str []string) bool {
	for i := 0; i < len(str)-1; i++ {
		switch {
		case str[i] == "(" && str[i+1] == ")":
			return false // проверка на то, чтобы скобки не закрывали друг друга без выражения внутри
		}
	}
	return true
}

// Calc вычисляет выражение и возвращает результат
func Calc(expression string) (float64, error) {
	tokens, err1 := tokenize(expression)
	if err1 != nil {
		return 0, err1
	}
	postfix, err2 := infixToPostfix(tokens)
	if err2 != nil {
		return 0, err2
	}
	result, err3 := evalPostfix(postfix)
	if err3 != nil {
		return 0, err3
	}
	return result, nil
}

// tokenize разбивает строку на токены (числа и операторы)
func tokenize(expression string) ([]string, error) {

	var tokens []string
	var current strings.Builder
	expression = strings.ReplaceAll(expression, " ", "")
	for i, char := range expression {
		if char >= '0' && char <= '9' || char == '.' {
			current.WriteRune(char)
		} else {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				if current.Len() > 1 && current.String()[0] == 48 {
					return nil, ErrInvalidSymbol
				}
				current.Reset()
			}
			if char == '+' || char == '-' || char == '*' || char == '/' || char == '(' || char == ')' {
				if char == '-' && i == 0 || (i > 0 && (expression[i-1] == '(')) {
					current.WriteRune(char) // добавляем знак минус к числу
				} else {
					tokens = append(tokens, string(char))
				}

			} else {
				return nil, ErrInvalidSymbol // проверка на недопустимый символ
			}
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	for _, v := range tokens {
		if v[0] == 48 && len(v) > 1 || len(v) > 1 && v[0] == 45 && v[1] == 48 {
			return nil, ErrInvalidSymbol // проверка на то, чтобы впереди числа не стоял 0
		}
	}

	if !CheckExpression(tokens) {
		return nil, ErrInvalidExpression
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
			output = append(output, token)
		} else if token == "(" {
			stack = append(stack, token)
		} else if token == ")" {
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return nil, ErrInvalidExpression // проверка на отсутствующую открывающая скобка
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
			return nil, ErrInvalidExpression // проверка на отсутствующую закрывающая скобка
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
			stack = append(stack, num)
		} else {
			if len(stack) < 2 {
				return 0, ErrInvalidExpression
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
