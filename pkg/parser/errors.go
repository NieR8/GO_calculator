package parser

import "fmt"

var (
	ErrInvalidSymbol     = fmt.Errorf("invalid symbol in expression")
	ErrInvalidExpression = fmt.Errorf("invalid expression")
	ErrEmptyExpression   = fmt.Errorf("empty expression")
	ErrInvalidRpn        = fmt.Errorf("invalid RPN expression")
	ErrDivisionByZero    = fmt.Errorf("division by zero")
)
