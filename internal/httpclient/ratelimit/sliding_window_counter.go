package ratelimit

import (
	"sync"
	"time"
)

type SlidingWindowCounter interface {
	Count() int64
	Increment()
}

// noLimitWindowCounter always returns 0 for Count and is used
// when a sliding window is not needed by the RateLimit
type noLimitWindowCounter struct{}

func (c *noLimitWindowCounter) Count() int64 {
	return 0
}

func (c *noLimitWindowCounter) Increment() {}

// bucket is used to keep track fo the number of requests a second
type bucket struct {
	// timestamp is a Unix timestamp in seconds
	timestamp int64
	// count is the number of requests at a timestamp
	count int64
}

// slidingWindowCounter is used to keep track of the number of requests
// in the configured windowSize
type slidingWindowCounter struct {
	lock sync.RWMutex

	// windowSize is the size of the window in seconds
	windowSize int64

	// buckets is used to keep track of the number of requests a second. The size
	// will always be equal to the windowSize
	buckets []bucket
}

// NewSlidingWindowCounter returns a slidingWindowCounter corresponding
// to the configured window size (in seconds). If a windowSize of zero
// is configured, a noLimitWindowCounter is returned.
func NewSlidingWindowCounter(windowSize int64) SlidingWindowCounter {
	if windowSize <= 0 {
		return &noLimitWindowCounter{}
	}
	return &slidingWindowCounter{
		windowSize: windowSize,
		buckets:    make([]bucket, windowSize),
	}
}

// Increment increments the counter for the current time.
func (c *slidingWindowCounter) Increment() {
	c.lock.Lock()
	defer c.lock.Unlock()

	now := time.Now().Unix()

	// Determine the bucket for the current second. buckets contains
	// an index for each second in windowSize
	index := now % c.windowSize
	b := &c.buckets[index]

	if b.timestamp == now {
		b.count++
	} else {
		b.timestamp = now
		b.count = 1
	}
}

// Count returns the total count in the sliding window.
func (c *slidingWindowCounter) Count() int64 {
	c.lock.RLock()
	defer c.lock.RUnlock()

	now := time.Now().Unix()

	// We only want the total count within the sliding window, so determine
	// the oldest bucket timestamp within the configured windowSize
	cutoff := now - c.windowSize

	total := int64(0)
	for _, b := range c.buckets {
		if b.timestamp > cutoff {
			total += b.count
		}
	}
	return total
}
