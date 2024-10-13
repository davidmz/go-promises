package promises

import (
	"errors"
	"sync"
)

// All takes an array of promises and returns a single promise. This returned
// promise fulfills when all of the input's promises fulfill (including when an
// empty iterable is passed), with an array of the fulfillment values. It
// rejects when any of the input's promises rejects, with this first rejection
// reason.
func All[T any](ps ...Promise[T]) Promise[[]T] {
	if len(ps) == 0 {
		return Resolve[[]T](nil)
	}
	return New(func() ([]T, error) {
		agg, abort := collectResults(ps)
		defer abort()

		values := make([]T, len(ps))
		settled := 0
		for r := range agg {
			settled++
			if r.Err != nil {
				return nil, r.Err
			}
			values[r.Index] = r.Value
			if settled == len(ps) {
				break
			}
		}

		return values, nil
	})
}

// Any takes an array of promises and returns a single promise. This returned
// promise fulfills when any of the input's promises fulfills, with this first
// fulfillment value. It rejects when all of the input's promises reject
// (including when an empty iterable is passed), with an [Errors]
// containing an array of rejection reasons.
func Any[T any](ps ...Promise[T]) Promise[T] {
	if len(ps) == 0 {
		return Reject[T](make(Errors, 0))
	}

	return New(func() (T, error) {
		agg, abort := collectResults(ps)
		defer abort()

		errs := make(Errors, len(ps))
		settled := 0
		for r := range agg {
			settled++
			if r.Err == nil {
				return r.Value, nil
			}
			errs[r.Index] = r.Err
			if settled == len(ps) {
				break
			}
		}

		return zero[T](), errs
	})
}

// Race takes an array of promises and returns a single Promise. This returned
// promise settles with the eventual state of the first promise that settles.
func Race[T any](ps ...Promise[T]) Promise[T] {
	if len(ps) == 0 {
		p, _, _ := WithResolvers[T]()
		return p
	}

	return New(func() (T, error) {
		agg, abort := collectResults(ps)
		defer abort()

		for r := range agg {
			return r.Value, r.Err
		}

		// We should never reach this
		return zero[T](), nil
	})
}

// AllSettled takes an array of promises and returns a single promise. This
// returned promise fulfills when all of the input's promises settle (including
// when an empty iterable is passed), with a [Results] objects that describe the
// outcome of each promise. Promise returned by this method is never rejected.
func AllSettled[T any](ps ...Promise[T]) Promise[Results[T]] {
	if len(ps) == 0 {
		return Resolve[Results[T]](nil)
	}

	return New(func() (Results[T], error) {
		agg, abort := collectResults(ps)
		defer abort()

		results := make(Results[T], len(ps))
		for r := range agg {
			results[r.Index] = r.Result
		}

		return results, nil
	})
}

// Result is an element of the [Results] array.
type Result[T any] struct {
	Value T
	Err   error
}

// Results represents the outcome of an resolved or rejected promise. It is
// used in the [AllSettled] response.
type Results[T any] []Result[T]

// Err returns all not-nil errors as a single error, or nil if there are no.
func (r Results[T]) Err() error {
	errs := make([]error, len(r))
	for i, result := range r {
		errs[i] = result.Err
	}
	return errors.Join(errs...)
}

type iResult[T any] struct {
	Index int
	Result[T]
}

func collectResults[T any](ps []Promise[T]) (resultsChan <-chan iResult[T], abort func()) {
	aggChan := make(chan iResult[T])
	abortChan := make(chan struct{})
	wg := new(sync.WaitGroup)
	wg.Add(len(ps))
	for i, p := range ps {
		go func(i int, p Promise[T]) {
			defer wg.Done()
			select {
			case <-p.Done():
			case <-abortChan:
				return
			}
			v, e := p.Wait()
			r := iResult[T]{i, Result[T]{v, e}}
			select {
			case aggChan <- r:
			case <-abortChan:
			}
		}(i, p)
	}
	go func() {
		wg.Wait()
		close(aggChan)
	}()
	return aggChan, func() { close(abortChan) }
}
