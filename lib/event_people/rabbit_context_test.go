package EventPeople

import (
	"os"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ----------------------------------------------------------------------------
// Test doubles
// ----------------------------------------------------------------------------

type publishedMsg struct {
	exchange string
	key      string
	msg      amqp.Publishing
}

type fakePublisher struct {
	published []publishedMsg
	err       error
}

func (f *fakePublisher) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	f.published = append(f.published, publishedMsg{exchange: exchange, key: key, msg: msg})
	return f.err
}

type fakeDelivery struct {
	acked       bool
	nacked      bool
	nackRequeue bool
}

func (d *fakeDelivery) Ack(multiple bool) error          { d.acked = true; return nil }
func (d *fakeDelivery) Nack(multiple, requeue bool) error { d.nacked = true; d.nackRequeue = requeue; return nil }
func (d *fakeDelivery) Reject(requeue bool) error         { return nil }

// ----------------------------------------------------------------------------
// Reject — app-level DLQ routing (spec §C: route to a plain <app>_dlq queue)
// ----------------------------------------------------------------------------

func TestReject_PublishesToDLQAndAcks(t *testing.T) {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "sophia")
	defer os.Unsetenv("RABBIT_EVENT_PEOPLE_APP_NAME")

	pub := &fakePublisher{}
	del := &fakeDelivery{}
	ctx := &RabbitContext{
		delivery: del,
		DLQName:  "sophia_dlq",
		channel:  pub,
	}
	ctx.DeliveryStruct.Body = []byte(`{"x":1}`)
	ctx.DeliveryStruct.ContentType = "application/json"

	ctx.Reject()

	if len(pub.published) != 1 {
		t.Fatalf("Reject: expected 1 publish to DLQ, got %d", len(pub.published))
	}
	p := pub.published[0]
	if p.exchange != "" {
		t.Errorf("Reject: expected default exchange \"\", got %q", p.exchange)
	}
	if p.key != "sophia_dlq" {
		t.Errorf("Reject: expected routing key sophia_dlq, got %q", p.key)
	}
	if string(p.msg.Body) != `{"x":1}` {
		t.Errorf("Reject: expected body forwarded, got %q", string(p.msg.Body))
	}
	if p.msg.ContentType != "application/json" {
		t.Errorf("Reject: expected content-type forwarded, got %q", p.msg.ContentType)
	}
	if !del.acked {
		t.Errorf("Reject: expected Ack after DLQ publish")
	}
	if del.nacked {
		t.Errorf("Reject: expected no Nack (DLX-based routing removed)")
	}
}

// ----------------------------------------------------------------------------
// Fail — exhausted retries route to the app-level DLQ (not Nack→DLX)
// ----------------------------------------------------------------------------

func TestFail_ExhaustedPublishesToDLQAndAcks(t *testing.T) {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "sophia")
	defer os.Unsetenv("RABBIT_EVENT_PEOPLE_APP_NAME")

	pub := &fakePublisher{}
	del := &fakeDelivery{}
	ctx := &RabbitContext{
		delivery:     del,
		DLQName:      "sophia_dlq",
		channel:      pub,
		MaxRetries:   3,
		initialDelay: 1000,
	}
	ctx.DeliveryStruct.Body = []byte(`{"x":1}`)
	ctx.DeliveryStruct.ContentType = "application/json"
	ctx.DeliveryStruct.QueueName = "sophia-resource.core.created.all"
	ctx.DeliveryStruct.DelayStrategy = "exponential"
	ctx.DeliveryStruct.RetryCount = 3 // == MaxRetries → exhausted

	ctx.Fail()

	if len(pub.published) != 1 {
		t.Fatalf("Fail (exhausted): expected 1 publish to DLQ, got %d", len(pub.published))
	}
	p := pub.published[0]
	if p.exchange != "" || p.key != "sophia_dlq" {
		t.Errorf("Fail (exhausted): expected publish to default exchange/sophia_dlq, got %q/%q", p.exchange, p.key)
	}
	if !del.acked {
		t.Errorf("Fail (exhausted): expected Ack after DLQ publish")
	}
	if del.nacked {
		t.Errorf("Fail (exhausted): expected no Nack (DLQ publish replaces Nack→DLX)")
	}
}

// Regression guard: with retries remaining, Fail still republishes to <queue>_retry.
func TestFail_WithRetriesRepublishesToRetryQueue(t *testing.T) {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "sophia")
	defer os.Unsetenv("RABBIT_EVENT_PEOPLE_APP_NAME")

	pub := &fakePublisher{}
	del := &fakeDelivery{}
	ctx := &RabbitContext{
		delivery:     del,
		DLQName:      "sophia_dlq",
		channel:      pub,
		MaxRetries:   3,
		initialDelay: 1000,
	}
	ctx.DeliveryStruct.Body = []byte(`{"x":1}`)
	ctx.DeliveryStruct.QueueName = "sophia-resource.core.created.all"
	ctx.DeliveryStruct.DelayStrategy = "exponential"
	ctx.DeliveryStruct.RetryCount = 0 // retries remain

	ctx.Fail()

	if len(pub.published) != 1 {
		t.Fatalf("Fail (retry): expected 1 publish to retry queue, got %d", len(pub.published))
	}
	p := pub.published[0]
	if p.key != "sophia-resource.core.created.all_retry" {
		t.Errorf("Fail (retry): expected publish to <queue>_retry, got %q", p.key)
	}
	if !del.acked {
		t.Errorf("Fail (retry): expected Ack after retry publish")
	}
}
