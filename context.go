package promises

import "context"

// Ctx creates a promise from a given context. It never resolves, and
// only rejects if the context is done.
func Ctx[T any](ctx context.Context) Promise[T] {
	if ctx.Err() != nil {
		return Reject[T](ctx.Err())
	}
	return New(func() (T, error) {
		<-ctx.Done()
		return zero[T](), ctx.Err()
	})
}

// WithContext creates a race between a given promise and a context. It is a
// shortcut for Race(promise, Ctx[T](ctx)).
func WithContext[T any](ctx context.Context, promise Promise[T]) Promise[T] {
	return Race(promise, Ctx[T](ctx))
}
