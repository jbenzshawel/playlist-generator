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

	counter := NewSlidingWindowCounter(window)

	go func() {
		for i := 0; i < 10; i++ {
			counter.Increment()
			time.Sleep(500 * time.Millisecond) //  2 requests per second
		}
	}()

	maxRequests := int64(window * 2) // window * 2 req / sec

	for idx := 0; idx < 5; idx++ {
		time.Sleep(1 * time.Second)

		assert.LessOrEqual(t, counter.Count(), maxRequests, "count greater than max requests in window")
	}
}
