package retry

import (
	"context"

	"github.com/sethvargo/go-retry"
)

func Do(ctx context.Context, b retry.Backoff, f retry.RetryFunc) error {
	return retry.Do(ctx, b, f)
}
