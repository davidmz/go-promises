package promises

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
		defer close(abort)

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
// (including when an empty iterable is passed), with an [AggregateError]
// containing an array of rejection reasons.
func Any[T any](ps ...Promise[T]) Promise[T] {
	if len(ps) == 0 {
		return Reject[T](new(AggregateError))
	}

	return New(func() (T, error) {
		agg, abort := collectResults(ps)
		defer close(abort)

		errs := make([]error, len(ps))
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

		return zero[T](), &AggregateError{errs}
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
		defer close(abort)

		for r := range agg {
			return r.Value, r.Err
		}

		// We should never reach this
		return zero[T](), nil
	})
}

// AllSettled takes an array of promises and returns a single promise. This
// returned promise fulfills when all of the input's promises settle (including
// when an empty iterable is passed), with an array of [Result] objects that
// describe the outcome of each promise.
func AllSettled[T any](ps ...Promise[T]) Promise[[]Result[T]] {
	if len(ps) == 0 {
		return Resolve[[]Result[T]](nil)
	}

	return New(func() ([]Result[T], error) {
		agg, abort := collectResults(ps)
		defer close(abort)

		results := make([]Result[T], len(ps))
		for r := range agg {
			results[r.Index] = r.Result
		}

		return results, nil
	})
}

// The result represents the outcome of an resolved or rejected promise. It is
// used in the [AllSettled] response.
type Result[T any] struct {
	Value T
	Err   error
}

type iResult[T any] struct {
	Index int
	Result[T]
}

func collectResults[T any](ps []Promise[T]) (<-chan iResult[T], chan<- struct{}) {
	agg := make(chan iResult[T])
	abort := make(chan struct{})
	for i, p := range ps {
		go func(i int, p Promise[T]) {
			select {
			case <-p.Done():
			case <-abort:
				return
			}
			v, e := p.Wait()
			r := iResult[T]{i, Result[T]{v, e}}
			select {
			case agg <- r:
			case <-abort:
			}
		}(i, p)
	}
	return agg, abort
}
