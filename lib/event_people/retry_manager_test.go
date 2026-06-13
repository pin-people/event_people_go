package EventPeople

import "testing"

func TestRetryManagerShouldRetry(t *testing.T) {
	rm := NewRetryManager(3, 1000, "exponential")

	if !rm.ShouldRetry() {
		t.Error("expected ShouldRetry=true on fresh RetryManager (attempt 0 < maxAttempts 3)")
	}

	rm.CurrentAttempt = 2
	if !rm.ShouldRetry() {
		t.Error("expected ShouldRetry=true when attempt=2 < maxAttempts=3")
	}

	rm.CurrentAttempt = 3
	if rm.ShouldRetry() {
		t.Error("expected ShouldRetry=false when attempt=3 >= maxAttempts=3")
	}
}

func TestRetryManagerExponentialDelay(t *testing.T) {
	rm := NewRetryManager(3, 1000, "exponential")

	// attempt 0: 1000 * 5^0 = 1000
	rm.CurrentAttempt = 0
	if rm.GetNextDelay() != 1000 {
		t.Errorf("expected 1000 at attempt 0, got %d", rm.GetNextDelay())
	}

	// attempt 1: 1000 * 5^1 = 5000
	rm.CurrentAttempt = 1
	if rm.GetNextDelay() != 5000 {
		t.Errorf("expected 5000 at attempt 1, got %d", rm.GetNextDelay())
	}

	// attempt 2: 1000 * 5^2 = 25000
	rm.CurrentAttempt = 2
	if rm.GetNextDelay() != 25000 {
		t.Errorf("expected 25000 at attempt 2, got %d", rm.GetNextDelay())
	}
}

func TestRetryManagerExponentialDelayCappedAtMaxDelay(t *testing.T) {
	rm := NewRetryManager(10, 1000, "exponential")
	rm.CurrentAttempt = 9 // 1000 * 5^9 = way over 600000
	delay := rm.GetNextDelay()
	if delay != MaxDelay {
		t.Errorf("expected delay capped at %d, got %d", MaxDelay, delay)
	}
}

func TestRetryManagerFixedDelay(t *testing.T) {
	rm := NewRetryManager(3, 2000, "fixed")

	rm.CurrentAttempt = 0
	if rm.GetNextDelay() != 2000 {
		t.Errorf("expected fixed delay 2000 at attempt 0, got %d", rm.GetNextDelay())
	}

	rm.CurrentAttempt = 2
	if rm.GetNextDelay() != 2000 {
		t.Errorf("expected fixed delay 2000 at attempt 2, got %d", rm.GetNextDelay())
	}
}

func TestRetryManagerIncrementAttempt(t *testing.T) {
	rm := NewRetryManager(3, 1000, "exponential")
	rm.IncrementAttempt()
	if rm.CurrentAttempt != 1 {
		t.Errorf("expected CurrentAttempt=1, got %d", rm.CurrentAttempt)
	}
}

func TestRetryManagerReset(t *testing.T) {
	rm := NewRetryManager(3, 1000, "exponential")
	rm.CurrentAttempt = 2
	rm.Reset()
	if rm.CurrentAttempt != 0 {
		t.Errorf("expected CurrentAttempt=0 after Reset, got %d", rm.CurrentAttempt)
	}
}
