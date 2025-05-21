package context

import (
	"context"
)

type Key interface {
	~struct{}
}

type ContextValue[keyType Key] interface {
	Key() keyType
}

func With[_type ContextValue[keyType], keyType Key](ctx context.Context, _ctx _type) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, _ctx.Key(), _ctx)
}

func From[_type ContextValue[keyType], keyType Key](ctx context.Context) (_type, bool) {
	var value _type
	_ctx, ok := ctx.Value(value.Key()).(_type)
	return _ctx, ok
}
