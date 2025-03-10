package promises

import (
	"errors"
	"fmt"
)

// ErrPanic returns from promise created by New or NewVoid when the generation
// function panics.
type ErrPanic struct {
	Value any
}

// Error returns the error text and makes ErrPanic compatible with the "error"
// interface.
func (p *ErrPanic) Error() string {
	return fmt.Sprintf("panic: %v", p.Value)
}

func handlePanic(reject func(error)) {
	if r := recover(); r != nil {
		reject(&ErrPanic{r})
	}
}

// Errors returns from [Any] function when some promises are rejected. It always
// contains the same number (and order) of errors as the number of promises
// passed. If some promise is fulfilled, the corresponding error is nil.
type Errors []error

// Err returns all not-nil errors as a single error.
func (e Errors) Err() error {
	return errors.Join(e...)
}

// Error returns the texts of all not-nil errors.
func (e Errors) Error() string {
	return e.Err().Error()
}
