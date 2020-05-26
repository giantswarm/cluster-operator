package controllercontext

import (
	"context"

	"github.com/giantswarm/microerror"
)

type key string

const contextKey key = "controller"

type Context struct {
	Client ContextClient
}

func NewContext(ctx context.Context, c Context) context.Context {
	return context.WithValue(ctx, contextKey, &c)
}

func FromContext(ctx context.Context) (*Context, error) {
	c, ok := ctx.Value(contextKey).(*Context)
	if !ok {
		return nil, microerror.Maskf(executionFailedError, "context key %q not found", contextKey)
	}

	return c, nil
}
