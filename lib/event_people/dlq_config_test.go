package EventPeople

import "testing"

func TestMaxRetries(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want int
	}{
		{"unset uses default", "", 3},
		{"valid value", "5", 5},
		{"garbage uses default", "abc", 3},
		{"zero uses default", "0", 3},
		{"negative uses default", "-2", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("RABBIT_EVENT_PEOPLE_MAX_RETRIES", tt.env)
			if got := MaxRetries(); got != tt.want {
				t.Errorf("MaxRetries() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRetryTTLMs(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want int64
	}{
		{"unset uses default", "", 30000},
		{"valid value", "5000", 5000},
		{"garbage uses default", "abc", 30000},
		{"zero uses default", "0", 30000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("RABBIT_EVENT_PEOPLE_RETRY_TTL_MS", tt.env)
			if got := RetryTTLMs(); got != tt.want {
				t.Errorf("RetryTTLMs() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDeadLetterNames(t *testing.T) {
	t.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "sophia")

	if got := DeadLetterExchangeName(); got != "sophia.dlx" {
		t.Errorf("DeadLetterExchangeName() = %q, want %q", got, "sophia.dlx")
	}
	if got := DeadLetterQueueName(); got != "sophia.dlq" {
		t.Errorf("DeadLetterQueueName() = %q, want %q", got, "sophia.dlq")
	}
	if got := RetryQueueName("sophia-resource.event.action.all"); got != "sophia-resource.event.action.all.retry" {
		t.Errorf("RetryQueueName() = %q, want %q", got, "sophia-resource.event.action.all.retry")
	}
}
