package promises_test

import (
	"errors"
	"testing"
	"time"

	"github.com/davidmz/go-promises"
	"github.com/stretchr/testify/suite"
)

func TestPromisesSuites(t *testing.T) {
	suite.Run(t, new(WithResolversSuite))
	suite.Run(t, new(ResolveRejectSuite))
	suite.Run(t, new(NewPromiseSuite))
}

type WithResolversSuite struct {
	suite.Suite
}

func (suite *WithResolversSuite) TestPromise() {
	promise, _, _ := promises.WithResolvers[int]()
	suite.False(isSettled(promise), "promise should not be settled")
}

func (suite *WithResolversSuite) TestResolve() {
	promise, resolve, _ := promises.WithResolvers[int]()
	suite.False(isSettled(promise), "promise should not be settled")
	resolve(42)
	suite.True(isSettled(promise), "promise should be settled")
	val, err := promise.Wait()
	suite.Equal(42, val, "promise should resolve with correct value")
	suite.Nil(err, "error should be nil")
}

func (suite *WithResolversSuite) TestReject() {
	promise, _, reject := promises.WithResolvers[int]()
	suite.False(isSettled(promise), "promise should not be settled")
	err := errors.New("some error")
	reject(err)
	suite.True(isSettled(promise), "promise should be settled")
	val, err1 := promise.Wait()
	suite.Equal(0, val, "promise value should be zero")
	suite.Equal(err, err1, "error should have the passed value")
}

func (suite *WithResolversSuite) TestResolveDelayed() {
	promise, resolve, _ := promises.WithResolvers[int]()
	time.Sleep(10 * time.Millisecond)
	suite.False(isSettled(promise), "promise should not be settled")
	resolve(42)
	suite.True(isSettled(promise), "promise should be settled")
}

func (suite *WithResolversSuite) TestResolveSecondCall() {
	promise, resolve, _ := promises.WithResolvers[int]()
	resolve(42)
	resolve(43)
	val, err := promise.Wait()
	suite.Equal(42, val, "promise should resolve with correct value")
	suite.Nil(err, "error should be nil")
}

func (suite *WithResolversSuite) TestRejectSecondCall() {
	promise, _, reject := promises.WithResolvers[int]()
	err := errors.New("some error")
	reject(err)
	reject(errors.New("some other error"))
	val, err1 := promise.Wait()
	suite.Equal(0, val, "promise value should be zero")
	suite.Equal(err, err1, "error should have the passed value")
}

type ResolveRejectSuite struct {
	suite.Suite
}

func (suite *ResolveRejectSuite) TestResolve() {
	promise := promises.Resolve(42)
	suite.True(isSettled(promise), "promise should be settled")
	val, err := promise.Wait()
	suite.Equal(42, val, "promise should resolve with correct value")
	suite.Nil(err, "error should be nil")
}

func (suite *ResolveRejectSuite) TestReject() {
	err := errors.New("some error")
	promise := promises.Reject[int](err)
	suite.True(isSettled(promise), "promise should be settled")
	val, err1 := promise.Wait()
	suite.Equal(0, val, "promise value should be zero")
	suite.Equal(err, err1, "error should have the passed value")
}

type NewPromiseSuite struct {
	suite.Suite
}

func (suite *NewPromiseSuite) TestNewResolve() {
	promise := promises.New(func() (int, error) { return 42, nil })
	suite.False(isSettled(promise), "promise should not be settled")
	val, err := promise.Wait()
	suite.True(isSettled(promise), "promise should be settled")
	suite.Equal(42, val, "promise should resolve with correct value")
	suite.Nil(err, "error should be nil")
}

func (suite *NewPromiseSuite) TestNewReject() {
	firedErr := errors.New("some error")
	promise := promises.New(func() (int, error) { return 42, firedErr })
	suite.False(isSettled(promise), "promise should not be settled")
	val, err := promise.Wait()
	suite.True(isSettled(promise), "promise should be settled")
	suite.Equal(0, val, "promise value should be zero")
	suite.Equal(firedErr, err, "error should have the passed value")
}

func (suite *NewPromiseSuite) TestNewPanic() {
	promise := promises.New(func() (int, error) { panic("AAA!") })
	suite.False(isSettled(promise), "promise should not be settled")
	val, err := promise.Wait()
	suite.True(isSettled(promise), "promise should be settled")
	suite.Equal(0, val, "promise value should be zero")
	suite.ErrorContains(err, "panic: AAA!")
}

func (suite *NewPromiseSuite) TestThen() {
	promise := promises.Resolve(42)
	promise2 := promises.Then(promise, func(x int) (int, error) { return x + 1, nil })
	val, err := promise2.Wait()
	suite.Equal(43, val, "promise should resolve with correct value")
	suite.Nil(err, "error should be nil")
}

func (suite *NewPromiseSuite) TestThenP() {
	promise := promises.Resolve(42)
	promise2 := promises.ThenP(promise, func(x int) promises.Promise[int] { return promises.Resolve(x + 1) })
	val, err := promise2.Wait()
	suite.Equal(43, val, "promise should resolve with correct value")
	suite.Nil(err, "error should be nil")
}
