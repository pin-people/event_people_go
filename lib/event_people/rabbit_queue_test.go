package EventPeople

import "testing"

func TestMainQueueArgs(t *testing.T) {
	t.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "sophia")

	args := mainQueueArgs()
	if got := args["x-dead-letter-exchange"]; got != "sophia.dlx" {
		t.Errorf("x-dead-letter-exchange = %v, want sophia.dlx", got)
	}
	if len(args) != 1 {
		t.Errorf("unexpected extra args: %v", args)
	}
}

func TestRetryQueueArgs(t *testing.T) {
	t.Setenv("RABBIT_EVENT_PEOPLE_RETRY_TTL_MS", "5000")

	args := retryQueueArgs("sophia-resource.event.action.all")
	if got := args["x-message-ttl"]; got != int64(5000) {
		t.Errorf("x-message-ttl = %v (%T), want 5000 (int64)", got, got)
	}
	if got := args["x-dead-letter-exchange"]; got != "" {
		t.Errorf("x-dead-letter-exchange = %v, want default exchange (empty)", got)
	}
	if got := args["x-dead-letter-routing-key"]; got != "sophia-resource.event.action.all" {
		t.Errorf("x-dead-letter-routing-key = %v, want original queue name", got)
	}
}
