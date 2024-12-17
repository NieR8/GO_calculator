package calculation

import "errors"

var (
	ErrInvalidExpression = errors.New("ошибка в выражении")
	ErrDivisionByZero    = errors.New("деление на 0")
	ErrInvalidOperation  = errors.New("недопустимая операция")
	ErrInvalidSymbol     = errors.New("недопустимый символ")
)
