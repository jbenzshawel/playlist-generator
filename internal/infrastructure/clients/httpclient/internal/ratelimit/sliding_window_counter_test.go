package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSlidingWindowCounter(t *testing.T) {
	const (
		window = 2
	)

	maxRequests := int64(window * 2) // window * 2 req / sec

	counter := NewSlidingWindowCounter(window)

	// Increment the counter 10 times in 5 seconds
	go func() {
		for i := 0; i < 10; i++ {
			counter.Increment()
			select {
			case <-t.Context().Done():
				return
			case <-time.After(500 * time.Millisecond): // simulate 2 requests per second
			}
		}
	}()

	// Verify the counter count never exceeds the maximum possible requests in the window
	for idx := 0; idx < 5; idx++ {
		time.Sleep(1 * time.Second)

		assert.LessOrEqual(t, counter.Count(), maxRequests, "count greater than max requests in window")
	}
}

func TestSlidingWindowCounter_ZeroWindowSize(t *testing.T) {
	counter := NewSlidingWindowCounter(0)

	for i := 0; i < 10; i++ {
		counter.Increment()
		assert.Zero(t, counter.Count())
	}
}
