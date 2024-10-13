// Package promises mimics the JavaScript Promise API, but in Go. You definitely
// don't need this package instead of ideomatic channels and goroutines, but
// sometimes and in the some cases it can be handy.
package promises

// Promise is a basic promise interface.
type Promise[T any] interface {
	// Wait waits for promise to settle and returns it value or error. If
	// promise is already settled, it returns immediately.
	Wait() (T, error)
	// Done returns a channel that is closed when the promise is settled. It is
	// useful for waiting promise with some other channels with "select".
	Done() <-chan struct{}
}

// WithResolvers returns a promise and two functions for resolve and reject it.
// After the first call to any of these functions, any subsequent calls will do
// nothing.
func WithResolvers[T any]() (
	promise Promise[T],
	resolve func(T),
	reject func(error),
) {
	p := &impl[T]{done: make(chan struct{})}
	return p, p.resolve, p.reject
}

// Resolve returns an already resolved Promise.
func Resolve[T any](value T) Promise[T] {
	p, resolve, _ := WithResolvers[T]()
	resolve(value)
	return p
}

// Reject returns an already rejected Promise.
func Reject[T any](err error) Promise[T] {
	p, _, reject := WithResolvers[T]()
	reject(err)
	return p
}

// New creates a promise that will be settled after the provided function
// returns. The function is called in the separate goroutine, so the New returns
// immediately, and the promise is settled asynchronously.
func New[T any](gen func() (T, error)) Promise[T] {
	p, resolve, reject := WithResolvers[T]()
	if gen == nil {
		resolve(zero[T]())
		return p
	}
	go func() {
		defer handlePanic(reject)
		value, err := gen()
		if err != nil {
			reject(err)
		} else {
			resolve(value)
		}
	}()
	return p
}

// NewVoid acting same as [New], but takes a function that returns only an error.
// It creates a promise with empty (struct{}) result.
func NewVoid(gen func() error) Promise[struct{}] {
	zero := struct{}{}
	if gen == nil {
		return Resolve(zero)
	}
	return New(func() (struct{}, error) { return zero, gen() })
}

// Then is an utility function that waits for the given promise and, if it
// fulfilled, processes the result using the gen function.
func Then[T, P any](p Promise[T], gen func(T) (P, error)) Promise[P] {
	return New((func() (P, error) {
		v, err := p.Wait()
		if err != nil {
			return zero[P](), err
		}
		return gen(v)
	}))
}

func zero[T any]() T { return *new(T) }
