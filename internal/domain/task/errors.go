package task

import "errors"

var (
	ErrNotFound          = errors.New("task: not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrAlreadyExists     = errors.New("task: already exists")
)