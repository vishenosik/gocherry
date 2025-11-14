package retry

import "github.com/sethvargo/go-retry"

func RetryableError(err error) error {
	return retry.RetryableError(err)
}
