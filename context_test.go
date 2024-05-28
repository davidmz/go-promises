package promises_test

import (
	"context"
	"testing"

	"github.com/davidmz/go-promises"
	"github.com/stretchr/testify/suite"
)

func TestContextSuite(t *testing.T) {
	suite.Run(t, new(ContextSuite))
}

type ContextSuite struct {
	suite.Suite
}

func (suite *ContextSuite) TestContext() {
	ctx, cancel := context.WithCancel(context.Background())
	promise := promises.Ctx[int](ctx)
	suite.False(isSettled(promise), "promise should not be settled")

	cancel()
	val, err := promise.Wait()

	suite.Equal(0, val, "promise value should be zero")
	suite.ErrorIs(err, context.Canceled, "error should be context.Canceled")
}

func (suite *ContextSuite) TestContext_canceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	promise := promises.Ctx[int](ctx)
	suite.True(isSettled(promise), "promise should be settled")

	val, err := promise.Wait()
	suite.Equal(0, val, "promise value should be zero")
	suite.ErrorIs(err, context.Canceled, "error should be context.Canceled")
}

func (suite *ContextSuite) TestWithContext_resolve() {
	promise, resolve, _ := promises.WithResolvers[int]()
	promise = promises.WithContext(context.Background(), promise)

	resolve(42)
	val, err := promise.Wait()
	suite.Equal(42, val, "promise should resolve with correct value")
	suite.Nil(err, "error should be nil")
}

func (suite *ContextSuite) TestWithContext_cancel() {
	ctx, cancel := context.WithCancel(context.Background())
	promise, _, _ := promises.WithResolvers[int]()

	promise = promises.WithContext(ctx, promise)
	cancel()

	val, err := promise.Wait()

	suite.Equal(0, val, "promise value should be zero")
	suite.ErrorIs(err, context.Canceled, "error should be context.Canceled")
}
