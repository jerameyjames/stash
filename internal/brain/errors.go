package brain

import "errors"

var (
	ErrEmptyContent = errors.New("brain: content cannot be empty")
	ErrInvalidLimit  = errors.New("brain: limit must be > 0")
	ErrNotFound      = errors.New("brain: memory not found")
	ErrInvalidID     = errors.New("brain: invalid memory ID")
)
