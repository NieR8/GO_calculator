package calculation

import "errors"

var (
	ErrInvalidExpression = errors.New("ошибка в выражении")
	ErrDivisionByZero    = errors.New("деление на 0")
	ErrNotEnoughArgs     = errors.New("недостаточно аргументов для операции")
	ErrInvalidOperation  = errors.New("недопустимая операция")
	ErrClosedParentheses = errors.New("отсутствует закрывающая скобка")
	ErrOpenedParentheses = errors.New("отсутствует открывающая скобка")
	ErrInvalidSymbol     = errors.New("недопустимый символ")
)
