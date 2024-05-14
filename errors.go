package razcache

import (
	"errors"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrWrongType = errors.New("wrong type")
)
