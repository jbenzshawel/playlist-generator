package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimit_SetLimited(t *testing.T) {
	t.Parallel()

	rl := New()

	d := 1 * time.Second

	rl.SetLimited(t.Context(), d)
	assert.True(t, rl.Limited())

	rl.WaitTimeAfter(t.Context())
	assert.False(t, rl.Limited())
}

func TestRateLimit_WithClientLimits_MaxRequestsLimited(t *testing.T) {
	t.Parallel()

	const (
		windowSize = 2
		maxReq     = 5
		batchSize  = 2
	)

	rl := New(WithClientLimits(windowSize, maxReq, batchSize))

	expectedLimitIndex := maxReq - batchSize

	for idx := range maxReq {
		rl.Increment(t.Context())

		expectLimited := idx >= expectedLimitIndex
		assert.Equal(t, expectLimited, rl.Limited())
	}

	rl.WaitTimeAfter(t.Context())
	assert.False(t, rl.Limited())
}

func TestRateLimit_WithClientLimits_NoBatch_MaxRequestsLimited(t *testing.T) {
	t.Parallel()

	const (
		windowSize = 2
		maxReq     = 5
		batchSize  = 0
	)

	rl := New(WithClientLimits(windowSize, maxReq, batchSize))

	for idx := range maxReq + 1 {
		rl.Increment(t.Context())

		expectLimited := idx >= maxReq
		assert.Equal(t, expectLimited, rl.Limited())
	}

	rl.WaitTimeAfter(t.Context())
	assert.False(t, rl.Limited())
}
