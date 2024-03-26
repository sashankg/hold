package main

import (
	"context"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
)

type LoggerExtension struct {
}

// ExecutionDidStart implements graphql.Extension.
func (*LoggerExtension) ExecutionDidStart(
	ctx context.Context,
) (context.Context, graphql.ExecutionFinishFunc) {
	return ctx, func(result *graphql.Result) {
		if result == nil {
			return
		}
		for _, err := range result.Errors {
			println("error", err.Error())
		}
	}

}

// GetResult implements graphql.Extension.
func (*LoggerExtension) GetResult(context.Context) interface{} {
	panic("unimplemented")
}

// HasResult implements graphql.Extension.
func (*LoggerExtension) HasResult() bool {
	return false
}

// Init implements graphql.Extension.
func (*LoggerExtension) Init(ctx context.Context, _ *graphql.Params) context.Context {
	return ctx
}

// Name implements graphql.Extension.
func (*LoggerExtension) Name() string {
	return "Logger"
}

// ParseDidStart implements graphql.Extension.
func (*LoggerExtension) ParseDidStart(
	ctx context.Context,
) (context.Context, graphql.ParseFinishFunc) {
	return ctx, func(error) {}
}

// ResolveFieldDidStart implements graphql.Extension.
func (*LoggerExtension) ResolveFieldDidStart(
	ctx context.Context,
	_ *graphql.ResolveInfo,
) (context.Context, graphql.ResolveFieldFinishFunc) {
	return ctx, func(any, error) {}
}

// ValidationDidStart implements graphql.Extension.
func (*LoggerExtension) ValidationDidStart(
	ctx context.Context,
) (context.Context, graphql.ValidationFinishFunc) {
	return ctx, func([]gqlerrors.FormattedError) {}
}

var _ graphql.Extension = &LoggerExtension{}
