package EventPeople

import (
	"testing"
)

// helper to build a RetryManager with a fixed initialDelay, independent of env vars.
func newRM(maxAttempts int, strategy string, initialDelay int) *RetryManager {
	return NewRetryManagerWithDelay(maxAttempts, strategy, initialDelay)
}

// ----------------------------------------------------------------------------
// ShouldRetry boundary tests
// ----------------------------------------------------------------------------

// retryCount == maxRetries-1  →  still within allowed retries, should retry.
func TestShouldRetry_AtLastAllowedRetry(t *testing.T) {
	rm := newRM(3, "exponential", 1000)
	retryCount := rm.MaxAttempts - 1 // 2
	if !rm.ShouldRetry(retryCount) {
		t.Errorf("ShouldRetry(%d) with MaxAttempts=%d: expected true, got false", retryCount, rm.MaxAttempts)
	}
}

// retryCount == maxRetries  →  exhausted, should NOT retry.
func TestShouldRetry_AtMaxRetries(t *testing.T) {
	rm := newRM(3, "exponential", 1000)
	retryCount := rm.MaxAttempts // 3
	if rm.ShouldRetry(retryCount) {
		t.Errorf("ShouldRetry(%d) with MaxAttempts=%d: expected false, got true", retryCount, rm.MaxAttempts)
	}
}

// ----------------------------------------------------------------------------
// GetNextDelay — exponential strategy
// ----------------------------------------------------------------------------

// retryCount=0 → initialDelay * 5^0 = 1000
func TestGetNextDelay_Exponential_RetryCount0(t *testing.T) {
	rm := newRM(5, "exponential", 1000)
	got := rm.GetNextDelay(0)
	want := 1000
	if got != want {
		t.Errorf("GetNextDelay(0) exponential: want %d, got %d", want, got)
	}
}

// retryCount=1 → initialDelay * 5^1 = 5000
func TestGetNextDelay_Exponential_RetryCount1(t *testing.T) {
	rm := newRM(5, "exponential", 1000)
	got := rm.GetNextDelay(1)
	want := 5000
	if got != want {
		t.Errorf("GetNextDelay(1) exponential: want %d, got %d", want, got)
	}
}

// retryCount=2 → initialDelay * 5^2 = 25000
func TestGetNextDelay_Exponential_RetryCount2(t *testing.T) {
	rm := newRM(5, "exponential", 1000)
	got := rm.GetNextDelay(2)
	want := 25000
	if got != want {
		t.Errorf("GetNextDelay(2) exponential: want %d, got %d", want, got)
	}
}

// ----------------------------------------------------------------------------
// GetNextDelay — cap at MaxDelay (600000 ms)
// ----------------------------------------------------------------------------

// A very large retryCount should be capped at 600000.
func TestGetNextDelay_Exponential_CappedAtMaxDelay(t *testing.T) {
	rm := newRM(100, "exponential", 1000)
	got := rm.GetNextDelay(20) // 1000 * 5^20 is astronomically large
	if got != rm.MaxDelay {
		t.Errorf("GetNextDelay(20) exponential: want cap %d, got %d", rm.MaxDelay, got)
	}
}

// retryCount=5 with initialDelay=1000: 1000*5^5 = 3_125_000 > 600_000 → capped.
func TestGetNextDelay_Exponential_ExceedsCapAtRetry5(t *testing.T) {
	rm := newRM(10, "exponential", 1000)
	got := rm.GetNextDelay(5)
	if got != rm.MaxDelay {
		t.Errorf("GetNextDelay(5) exponential: want cap %d, got %d", rm.MaxDelay, got)
	}
}

// ----------------------------------------------------------------------------
// GetNextDelay — fixed strategy
// ----------------------------------------------------------------------------

// Fixed strategy always returns initialDelay, regardless of retryCount.
func TestGetNextDelay_Fixed_AlwaysReturnsInitialDelay(t *testing.T) {
	rm := newRM(5, "fixed", 2500)
	for _, rc := range []int{0, 1, 2, 10} {
		got := rm.GetNextDelay(rc)
		if got != rm.InitialDelay {
			t.Errorf("GetNextDelay(%d) fixed: want %d, got %d", rc, rm.InitialDelay, got)
		}
	}
}
