package ratelimit

import (
	"sync"
	"time"
)

type windowCounter interface {
	Count() int64
	Increment()
}

type noLimitWindowCounter struct{}

func (c *noLimitWindowCounter) Count() int64 {
	return 0
}

func (c *noLimitWindowCounter) Increment() {}

type bucket struct {
	// timestamp is a Unix timestamp in seconds
	timestamp int64
	count     int64
}

type slidingWindowCounter struct {
	lock sync.RWMutex

	windowSize int64 // Window size in seconds
	buckets    []bucket
}

func NewSlidingWindowCounter(windowSizeInSeconds int64) *slidingWindowCounter {
	if windowSizeInSeconds <= 0 {
		windowSizeInSeconds = 60 // Default to 60 seconds if invalid
	}
	return &slidingWindowCounter{
		windowSize: windowSizeInSeconds,
		buckets:    make([]bucket, windowSizeInSeconds),
	}
}

// Increment increments the counter for the current time.
func (c *slidingWindowCounter) Increment() {
	c.lock.Lock()
	defer c.lock.Unlock()

	now := time.Now().Unix()

	// Get the bucket index for the current second
	index := now % c.windowSize

	b := &c.buckets[index]

	if b.timestamp == now {
		b.count++
	} else {
		b.timestamp = now
		b.count = 1
	}
}

// Count returns the total count of requests in the sliding window.
func (c *slidingWindowCounter) Count() int64 {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var total int64 = 0
	now := time.Now().Unix()

	cutoff := now - c.windowSize

	for _, b := range c.buckets {
		if b.timestamp > cutoff {
			total += b.count
		}
	}
	return total
}
