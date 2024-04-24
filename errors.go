package promises

import (
	"fmt"
	"strings"
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

// AggregateError returns from [Any] function when some promises are rejected.
// Its Errors field always returns the same number (and order) of errors as the
// number of promises passed. If some promise is fulfilled, the corresponding
// error is nil.
type AggregateError struct {
	Errors []error
}

// Error returns the "\n"-join of all not-nil errors.
func (e *AggregateError) Error() string {
	var b strings.Builder
	for _, err := range e.Errors {
		if err == nil {
			continue
		}
		if b.Len() > 0 {
			b.WriteRune('\n')
		}
		b.WriteString(err.Error())
	}
	if b.Len() == 0 {
		b.WriteString("empty error")
	}
	return b.String()
}
