package EventPeople

import (
	"math"
	"os"
	"strconv"
)

// RetryManager handles retry logic including delay calculation and retry eligibility.
type RetryManager struct {
	MaxAttempts   int
	DelayStrategy string
	InitialDelay  int // from RABBIT_EVENT_PEOPLE_RETRY_TTL_MS, default 1000
	MaxDelay      int // 600000ms = 10 minutes
}

// NewRetryManager creates a RetryManager with the given configuration.
// InitialDelay is read from RABBIT_EVENT_PEOPLE_RETRY_TTL_MS (default 1000).
func NewRetryManager(maxAttempts int, delayStrategy string) *RetryManager {
	initialDelay := 1000
	if v := os.Getenv("RABBIT_EVENT_PEOPLE_RETRY_TTL_MS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			initialDelay = parsed
		}
	}
	return &RetryManager{
		MaxAttempts:   maxAttempts,
		DelayStrategy: delayStrategy,
		InitialDelay:  initialDelay,
		MaxDelay:      600000,
	}
}

// ShouldRetry returns true if retryCount is less than MaxAttempts.
func (r *RetryManager) ShouldRetry(retryCount int) bool {
	return retryCount < r.MaxAttempts
}

// GetNextDelay returns the delay in milliseconds for the next retry attempt.
// Exponential: min(initialDelay * 5^retryCount, maxDelay)
// Fixed: initialDelay
func (r *RetryManager) GetNextDelay(retryCount int) int {
	if r.DelayStrategy == "fixed" {
		return r.InitialDelay
	}
	// exponential (default)
	delay := float64(r.InitialDelay) * math.Pow(5, float64(retryCount))
	if delay > float64(r.MaxDelay) {
		return r.MaxDelay
	}
	return int(delay)
}
