package cache

import (
	// std
	"context"
	"time"
	//pkg
)

type NoopCache struct{}

func NewNoopCache() CacheProvider { return newNoopCache() }

func newNoopCache() *NoopCache { return new(NoopCache) }

func (noop *NoopCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return nil
}

func (noop *NoopCache) Get(ctx context.Context, key string) (string, error) { return "", nil }

func (noop *NoopCache) Delete(ctx context.Context, key string) error { return nil }

func (noop *NoopCache) Close(ctx context.Context) error { return nil }
