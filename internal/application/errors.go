package application

import "errors"

var (
	ErrInvalidExprName = errors.New("в теле запроса ключ выражения должен называться expression либо выражение не должно быть пустым")
	ErrInternalServer  = errors.New("internal server error")
)
