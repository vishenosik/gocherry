package retry

import (
	"math"
	"sync/atomic"
	"time"
	"unsafe"
)

const defaultBase = time.Second * 10

type state [2]time.Duration

type fibonacciBackoff struct {
	state       unsafe.Pointer
	maxDuration time.Duration
}

// NewFibonacci creates a new Fibonacci backoff using the starting value of
// base. The wait time is the sum of the previous two wait times on each failed
// attempt (1, 1, 2, 3, 5, 8, 13...).
//
// Once it overflows, the function constantly returns the maximum time.Duration
// for a 64-bit integer.
func NewFibonacci(base time.Duration, maxDuration time.Duration) *fibonacciBackoff {
	if base <= 0 {
		base = defaultBase
	}

	if maxDuration <= 0 {
		maxDuration = math.MaxInt64
	}

	return &fibonacciBackoff{
		state:       unsafe.Pointer(&state{0, base}),
		maxDuration: maxDuration,
	}
}

// Next implements Backoff. It is safe for concurrent use.
func (b *fibonacciBackoff) Next() (next time.Duration, stop bool) {
	for {
		curr := atomic.LoadPointer(&b.state)
		currState := (*state)(curr)
		nx := b.next()

		if atomic.CompareAndSwapPointer(&b.state, curr, unsafe.Pointer(&state{currState[1], nx})) {
			return nx, false
		}
	}
}

func (b *fibonacciBackoff) RetryInSeconds() int64 {
	return int64(b.next() / time.Second)
}

func (b *fibonacciBackoff) next() time.Duration {
	curr := atomic.LoadPointer(&b.state)
	currState := (*state)(curr)
	n := currState[0] + currState[1]

	if n >= b.maxDuration {
		return b.maxDuration
	}
	return n
}
