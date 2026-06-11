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
// Prefer NewRetryManagerWithDelay when the caller has already resolved the delay
// (e.g. stored in RabbitContext) to avoid repeated env-var lookups.
func NewRetryManager(maxAttempts int, delayStrategy string) *RetryManager {
	return NewRetryManagerWithDelay(maxAttempts, delayStrategy, resolveInitialDelay())
}

// NewRetryManagerWithDelay creates a RetryManager with a pre-resolved initialDelay.
func NewRetryManagerWithDelay(maxAttempts int, delayStrategy string, initialDelay int) *RetryManager {
	return &RetryManager{
		MaxAttempts:   maxAttempts,
		DelayStrategy: delayStrategy,
		InitialDelay:  initialDelay,
		MaxDelay:      600000,
	}
}

// resolveInitialDelay reads RABBIT_EVENT_PEOPLE_RETRY_TTL_MS once and returns the value.
func resolveInitialDelay() int {
	initialDelay := 1000
	if v := os.Getenv("RABBIT_EVENT_PEOPLE_RETRY_TTL_MS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			initialDelay = parsed
		}
	}
	return initialDelay
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
