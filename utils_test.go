package promises_test

import "github.com/davidmz/go-promises"

func isSettled[T any](promise promises.Promise[T]) bool {
	select {
	case <-promise.Done():
		return true
	default:
		return false
	}
}
