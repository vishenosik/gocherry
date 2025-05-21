package context

import (
	"context"
	"os"
)

type stopContextKey struct{}

type stopContext struct {
	Signal os.Signal
}

func (ctx *stopContext) Key() stopContextKey {
	return stopContextKey{}
}

func WithStopCtx(ctx context.Context, signal os.Signal) context.Context {
	return With(ctx, &stopContext{
		Signal: signal,
	})
}

func StopFromCtx(ctx context.Context) (*stopContext, bool) {
	return From[*stopContext](ctx)
}
