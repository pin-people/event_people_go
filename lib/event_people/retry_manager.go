package EventPeople

import "math"

const (
	// MaxDelay is the maximum retry delay in milliseconds (10 minutes).
	MaxDelay = 600000
)

// RetryManager manages retry policies and dead-lettering for a single listener.
// It is an internal component — not exposed directly to users.
type RetryManager struct {
	MaxAttempts    int
	InitialDelay   int
	DelayStrategy  string
	CurrentAttempt int
}

// NewRetryManager creates a RetryManager with the given configuration.
func NewRetryManager(maxAttempts int, initialDelay int, delayStrategy string) *RetryManager {
	return &RetryManager{
		MaxAttempts:   maxAttempts,
		InitialDelay:  initialDelay,
		DelayStrategy: delayStrategy,
	}
}

// ShouldRetry returns true if CurrentAttempt < MaxAttempts.
func (rm *RetryManager) ShouldRetry() bool {
	return rm.CurrentAttempt < rm.MaxAttempts
}

// GetNextDelay calculates the next delay in milliseconds.
// Exponential: min(initialDelay * (5 ^ currentAttempt), maxDelay).
// Fixed: initialDelay (constant).
func (rm *RetryManager) GetNextDelay() int {
	if rm.DelayStrategy == "fixed" {
		return rm.InitialDelay
	}
	// exponential: initialDelay * (5 ^ currentAttempt)
	delay := float64(rm.InitialDelay) * math.Pow(5, float64(rm.CurrentAttempt))
	if delay > MaxDelay {
		return MaxDelay
	}
	return int(delay)
}

// IncrementAttempt increments the current attempt counter.
func (rm *RetryManager) IncrementAttempt() {
	rm.CurrentAttempt++
}

// Reset resets CurrentAttempt to 0 for a new event.
func (rm *RetryManager) Reset() {
	rm.CurrentAttempt = 0
}
