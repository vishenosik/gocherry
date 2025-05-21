package context

import "context"

type requestContextKey struct{}

type requestContext struct {
	requestID string
}

func (ctx *requestContext) Key() requestContextKey {
	return requestContextKey{}
}

func WithRequestCtx(ctx context.Context, requestID string) context.Context {
	return With(ctx, &requestContext{
		requestID: requestID,
	})
}

func RequestFromCtx(ctx context.Context) (*requestContext, bool) {
	return From[*requestContext](ctx)
}
