package EventPeople

import (
	"context"
	"errors"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

// mockRetryPublisher satisfies retryPublisher and records what was published.
type mockRetryPublisher struct {
	published []amqp.Publishing
	keys      []string
	failWith  error
}

func (m *mockRetryPublisher) PublishWithContext(
	ctx context.Context,
	exchange, key string,
	mandatory, immediate bool,
	msg amqp.Publishing,
) error {
	if m.failWith != nil {
		return m.failWith
	}
	m.published = append(m.published, msg)
	m.keys = append(m.keys, key)
	return nil
}

// mockDelivery is a test double that records ack/nack/reject calls.
type mockDelivery struct {
	acked    bool
	nacked   bool
	rejected bool
	requeue  bool
}

func (m *mockDelivery) Ack(multiple bool) error {
	m.acked = true
	return nil
}

func (m *mockDelivery) Nack(multiple bool, requeue bool) error {
	m.nacked = true
	m.requeue = requeue
	return nil
}

func (m *mockDelivery) Reject(requeue bool) error {
	m.rejected = true
	m.requeue = requeue
	return nil
}

// failingDelivery always returns errors (simulates broker down).
type failingDelivery struct {
	nacked  bool
	requeue bool
}

func (f *failingDelivery) Ack(multiple bool) error {
	return errors.New("broker down")
}

func (f *failingDelivery) Nack(multiple bool, requeue bool) error {
	f.nacked = true
	f.requeue = requeue
	return nil
}

func (f *failingDelivery) Reject(requeue bool) error {
	return errors.New("broker down")
}

func TestRabbitContextSuccess(t *testing.T) {
	d := &mockDelivery{}
	ctx := NewContext(d)
	ctx.Success()
	if !d.acked {
		t.Error("expected Ack to be called on Success")
	}
}

// Reject with no channel/DLQ configured falls back to nack(requeue=false).
func TestRabbitContextRejectNoChannelNacks(t *testing.T) {
	d := &mockDelivery{}
	ctx := NewContext(d)
	ctx.Reject()
	if !d.nacked || d.requeue {
		t.Error("expected Nack(requeue=false) fallback when no channel/DLQ is configured")
	}
}

// Reject routes the message to the application-level DLQ (publish + ack), not a
// broker dead-letter-exchange.
func TestRabbitContextRejectPublishesToDLQ(t *testing.T) {
	d := &mockDelivery{}
	pub := &mockRetryPublisher{}
	ctx := NewContextWithRetry(d, pub, "myqueue_retry", 0, 3, "myapp_dlq", 1000, "exponential")
	ctx.DeliveryStruct = DeliveryStruct{Body: []byte(`{"x":1}`), ContentType: "application/json"}

	ctx.Reject()

	if len(pub.published) != 1 || pub.keys[0] != "myapp_dlq" {
		t.Fatalf("Reject: expected 1 publish to myapp_dlq, got %d keys=%v", len(pub.published), pub.keys)
	}
	if string(pub.published[0].Body) != `{"x":1}` {
		t.Errorf("Reject: expected body forwarded to DLQ, got %q", string(pub.published[0].Body))
	}
	if !d.acked || d.nacked {
		t.Error("Reject: expected Ack after DLQ publish, no Nack")
	}
}

// Retry exhaustion (with a channel) routes the message to the DLQ via publish + ack.
func TestRabbitContextFailExhaustedPublishesToDLQ(t *testing.T) {
	d := &mockDelivery{}
	pub := &mockRetryPublisher{}
	// retryCount=3, maxRetries=3 → exhausted; channel present → publish to DLQ.
	ctx := NewContextWithRetry(d, pub, "myqueue_retry", 3, 3, "myapp_dlq", 1000, "exponential")
	ctx.DeliveryStruct = DeliveryStruct{Body: []byte(`{}`)}

	ctx.Fail()

	if len(pub.published) != 1 || pub.keys[0] != "myapp_dlq" {
		t.Fatalf("exhausted: expected 1 publish to myapp_dlq, got %d keys=%v", len(pub.published), pub.keys)
	}
	if !d.acked || d.nacked {
		t.Error("exhausted: expected Ack after DLQ publish, no Nack")
	}
}

func TestRabbitContextFailNoChannel(t *testing.T) {
	// When no AMQP channel is set, Fail must nack(requeue=false).
	d := &mockDelivery{}
	ctx := NewContext(d)
	ctx.Fail()
	if !d.nacked {
		t.Error("expected Nack to be called when no channel is configured")
	}
	if d.requeue {
		t.Error("expected requeue=false when no channel is configured — must not cause infinite loop")
	}
}

func TestRabbitContextFailRetriesExhausted(t *testing.T) {
	// retryCount=3, maxRetries=3 → retries exhausted → nack(requeue=false).
	d := &mockDelivery{}
	ctx := NewContextWithRetry(d, nil, "myqueue_retry", 3, 3, "myapp_dlq", 1000, "exponential")
	ctx.Fail()
	if !d.nacked {
		t.Error("expected Nack when retries exhausted")
	}
	if d.requeue {
		t.Error("expected requeue=false when retries exhausted")
	}
}

func TestRabbitContextIsLastRetry(t *testing.T) {
	// retryCount=2, maxRetries=3 → isLastRetry = (2 >= 3-1) = true
	d := &mockDelivery{}
	ctx := NewContextWithRetry(d, nil, "myqueue_retry", 2, 3, "myapp_dlq", 1000, "exponential")
	if !ctx.IsLastRetry {
		t.Error("expected IsLastRetry=true when retryCount >= maxRetries-1")
	}
}

func TestRabbitContextIsNotLastRetry(t *testing.T) {
	// retryCount=0, maxRetries=3 → isLastRetry = (0 >= 2) = false
	d := &mockDelivery{}
	ctx := NewContextWithRetry(d, nil, "myqueue_retry", 0, 3, "myapp_dlq", 1000, "exponential")
	if ctx.IsLastRetry {
		t.Error("expected IsLastRetry=false when retryCount=0 and maxRetries=3")
	}
}

func TestRabbitContextGetMaxRetries(t *testing.T) {
	d := &mockDelivery{}
	ctx := NewContextWithRetry(d, nil, "myqueue_retry", 0, 5, "app_dlq", 1000, "exponential")
	if ctx.GetMaxRetries() != 5 {
		t.Errorf("expected GetMaxRetries()=5, got %d", ctx.GetMaxRetries())
	}
}

func TestRabbitContextGetIsLastRetry(t *testing.T) {
	d := &mockDelivery{}
	ctx := NewContextWithRetry(d, nil, "myqueue_retry", 2, 3, "app_dlq", 1000, "exponential")
	if !ctx.GetIsLastRetry() {
		t.Error("expected GetIsLastRetry()=true")
	}
}

// TestRabbitContextFailUsesListenerInitialDelay verifies that the listener-resolved
// initialDelay (not the global Config value) is used when calculating the retry delay.
// This is the regression test for the bug where initialDelay was always read from
// Config.InitialDelay, silently dropping any per-listener override.
func TestRabbitContextFailUsesListenerInitialDelay(t *testing.T) {
	// Use a listener-level initialDelay of 500ms (fixed strategy, attempt 0 → delay = 500).
	// The global Config default is 1000ms. If the bug is present, the expiration would be
	// "1000"; if fixed, it will be "500".
	listenerInitialDelay := 500
	listenerDelayStrategy := "fixed"

	d := &mockDelivery{}
	pub := &mockRetryPublisher{}

	ctx := NewContextWithRetry(
		d,
		pub,
		"myqueue_retry",
		0,             // retryCount=0 → retries remain (maxAttempts=3)
		3,             // maxAttempts
		"myapp_dlq",
		listenerInitialDelay,
		listenerDelayStrategy,
	)
	ctx.DeliveryStruct = DeliveryStruct{Body: []byte(`{}`)}

	ctx.Fail()

	if len(pub.published) != 1 {
		t.Fatalf("expected exactly 1 message published to retry queue, got %d", len(pub.published))
	}

	got := pub.published[0].Expiration
	want := "500"
	if got != want {
		t.Errorf("Fail() used wrong initialDelay: Expiration=%q, want %q — listener initialDelay override was silently dropped", got, want)
	}

	if !d.acked {
		t.Error("expected Ack after successful publish to retry queue")
	}
}

// TestRabbitContextFailUsesListenerDelayStrategy verifies that the listener-resolved
// delayStrategy is used when calculating the retry delay.
func TestRabbitContextFailUsesListenerDelayStrategy(t *testing.T) {
	// exponential at attempt 0: initialDelay * 5^0 = 2000 * 1 = 2000
	// fixed at attempt 0: initialDelay = 2000
	// Use exponential to distinguish from fixed: at attempt 1, exponential = 2000*5 = 10000,
	// fixed = 2000. Test at attempt 1 with exponential strategy.
	d := &mockDelivery{}
	pub := &mockRetryPublisher{}

	ctx := NewContextWithRetry(
		d,
		pub,
		"myqueue_retry",
		1,             // retryCount=1 → retries remain (maxAttempts=3)
		3,             // maxAttempts
		"myapp_dlq",
		2000,          // initialDelay
		"exponential", // delayStrategy — exponential at attempt 1: 2000 * 5^1 = 10000
	)
	ctx.DeliveryStruct = DeliveryStruct{Body: []byte(`{}`)}

	ctx.Fail()

	if len(pub.published) != 1 {
		t.Fatalf("expected exactly 1 message published to retry queue, got %d", len(pub.published))
	}

	got := pub.published[0].Expiration
	want := "10000"
	if got != want {
		t.Errorf("Fail() used wrong delayStrategy: Expiration=%q, want %q — listener delayStrategy override was silently dropped", got, want)
	}
}

// TestRabbitContextFailPublishErrorNacksWithoutRequeue verifies that a publish failure
// results in nack(requeue=false) — never requeue=true — to avoid infinite redelivery loops.
func TestRabbitContextFailPublishErrorNacksWithoutRequeue(t *testing.T) {
	d := &mockDelivery{}
	pub := &mockRetryPublisher{failWith: errors.New("broker unavailable")}

	ctx := NewContextWithRetry(
		d,
		pub,
		"myqueue_retry",
		0,
		3,
		"myapp_dlq",
		1000,
		"exponential",
	)
	ctx.DeliveryStruct = DeliveryStruct{Body: []byte(`{}`)}

	ctx.Fail()

	if !d.nacked {
		t.Error("expected Nack when publish to retry queue fails")
	}
	if d.requeue {
		t.Error("expected requeue=false when publish to retry queue fails — requeue=true causes infinite loop")
	}
	if d.acked {
		t.Error("must not Ack when publish to retry queue fails")
	}
}
