package promises_test

import (
	"errors"
	"testing"
	"time"

	"github.com/davidmz/go-promises"
	"github.com/stretchr/testify/suite"
)

func TestAggregatesSuite(t *testing.T) {
	suite.Run(t, new(AggregatesSuite))
}

type AggregatesSuite struct {
	suite.Suite
}

// All

func (suite *AggregatesSuite) TestAll_empty() {
	promise := promises.All[int]()
	suite.True(isSettled(promise), "promise should be settled")
	val, err := promise.Wait()
	suite.Nil(val)
	suite.Nil(err)
}

func (suite *AggregatesSuite) TestAll_all_resolved() {
	p := promises.All(
		promises.Resolve(41),
		promises.Resolve(42),
		promises.Resolve(43),
	)
	val, err := p.Wait()
	suite.Equal([]int{41, 42, 43}, val)
	suite.Nil(err)
}

func (suite *AggregatesSuite) TestAll_one_rejected() {
	tgtErr := errors.New("test error")
	p := promises.All(
		promises.Resolve(41),
		promises.Reject[int](tgtErr),
		promises.Resolve(43),
	)
	val, err := p.Wait()
	suite.Nil(val)
	suite.Equal(tgtErr, err)
}

func (suite *AggregatesSuite) TestAll_all_rejected() {
	tgtErr1 := errors.New("test error 1")
	tgtErr2 := errors.New("test error 2")
	tgtErr3 := errors.New("test error 3")
	p := promises.All(
		promises.Reject[int](tgtErr1),
		promises.Reject[int](tgtErr2),
		promises.Reject[int](tgtErr3),
	)
	val, err := p.Wait()
	suite.Nil(val)
	suite.Contains([]error{tgtErr1, tgtErr2, tgtErr3}, err)
}

func (suite *AggregatesSuite) TestAll_delayed_resolve() {
	p1, resolve1, _ := promises.WithResolvers[int]()
	p2, resolve2, _ := promises.WithResolvers[int]()

	promise := promises.All(p1, p2)
	time.Sleep(10 * time.Millisecond)
	suite.False(isSettled(promise), "promise should not be settled")

	resolve1(42)
	time.Sleep(10 * time.Millisecond)
	suite.False(isSettled(promise), "promise should not be settled")

	resolve2(43)
	time.Sleep(10 * time.Millisecond)
	suite.True(isSettled(promise), "promise should be settled")

	val, err := promise.Wait()
	suite.Equal([]int{42, 43}, val)
	suite.Nil(err)
}

func (suite *AggregatesSuite) TestAll_delayed_reject() {
	tgtErr := errors.New("test error")
	p1, _, reject1 := promises.WithResolvers[int]()
	p2, resolve2, _ := promises.WithResolvers[int]()

	promise := promises.All(p1, p2)
	time.Sleep(10 * time.Millisecond)
	suite.False(isSettled(promise), "promise should not be settled")

	reject1(tgtErr)
	time.Sleep(10 * time.Millisecond)
	suite.True(isSettled(promise), "promise should be settled")
	{
		val, err := promise.Wait()
		suite.Nil(val)
		suite.Equal(tgtErr, err)
	}

	resolve2(43)
	time.Sleep(10 * time.Millisecond)

	{
		val, err := promise.Wait()
		suite.Nil(val)
		suite.Equal(tgtErr, err)
	}
}

// Any

func (suite *AggregatesSuite) TestAny_empty() {
	promise := promises.Any[int]()
	suite.True(isSettled(promise), "promise should be settled")
	val, err := promise.Wait()
	suite.Zero(val)
	var expectedErr *promises.AggregateError
	suite.ErrorAs(err, &expectedErr)
	suite.Empty(expectedErr.Errors)
}

func (suite *AggregatesSuite) TestAny_all_resolved() {
	p := promises.Any(
		promises.Resolve(41),
		promises.Resolve(42),
		promises.Resolve(43),
	)
	val, err := p.Wait()
	suite.Contains([]int{41, 42, 43}, val)
	suite.Nil(err)
}

func (suite *AggregatesSuite) TestAny_all_rejected() {
	tgtErr1 := errors.New("test error 1")
	tgtErr2 := errors.New("test error 2")
	tgtErr3 := errors.New("test error 3")
	p := promises.Any(
		promises.Reject[int](tgtErr1),
		promises.Reject[int](tgtErr2),
		promises.Reject[int](tgtErr3),
	)
	val, err := p.Wait()
	suite.Zero(val)
	var expectedErr *promises.AggregateError
	suite.ErrorAs(err, &expectedErr)
	suite.Equal([]error{tgtErr1, tgtErr2, tgtErr3}, expectedErr.Errors)
}

func (suite *AggregatesSuite) TestAny_one_resolved() {
	p := promises.Any(
		promises.Reject[int](errors.New("test error 1")),
		promises.Resolve(42),
		promises.Reject[int](errors.New("test error 3")),
	)
	val, err := p.Wait()
	suite.Equal(42, val)
	suite.Nil(err)
}

func (suite *AggregatesSuite) TestAny_delayed() {
	p1, _, reject1 := promises.WithResolvers[int]()
	p2, resolve2, _ := promises.WithResolvers[int]()

	promise := promises.Any(p1, p2)
	time.Sleep(10 * time.Millisecond)
	suite.False(isSettled(promise), "promise should not be settled")

	reject1(errors.New("some error"))
	time.Sleep(10 * time.Millisecond)
	suite.False(isSettled(promise), "promise should not be settled")

	resolve2(42)
	time.Sleep(10 * time.Millisecond)
	suite.True(isSettled(promise), "promise should be settled")

	val, err := promise.Wait()
	suite.Equal(42, val)
	suite.Nil(err)
}

func (suite *AggregatesSuite) TestAllSettled() {
	p := promises.AllSettled(
		promises.Resolve(41),
		promises.Resolve(42),
		promises.Resolve(43),
	)
	val, err := p.Wait()
	suite.Equal([]promises.Result[int]{
		{41, nil},
		{42, nil},
		{43, nil},
	}, val)
	suite.Nil(err)
}
