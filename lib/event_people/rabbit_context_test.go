package EventPeople

import (
	"errors"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

type fakeDelivery struct {
	acks    int
	nacks   []bool // requeue flag per Nack call
	rejects []bool // requeue flag per Reject call
}

func (f *fakeDelivery) Ack(bool) error { f.acks++; return nil }
func (f *fakeDelivery) Nack(_ bool, requeue bool) error {
	f.nacks = append(f.nacks, requeue)
	return nil
}
func (f *fakeDelivery) Reject(requeue bool) error {
	f.rejects = append(f.rejects, requeue)
	return nil
}

type fakePublisher struct {
	routingKeys []string
	headers     []amqp.Table
	bodies      [][]byte
	err         error
}

func (f *fakePublisher) publish(routingKey string, headers amqp.Table, body []byte) error {
	if f.err != nil {
		return f.err
	}
	f.routingKeys = append(f.routingKeys, routingKey)
	f.headers = append(f.headers, headers)
	f.bodies = append(f.bodies, body)
	return nil
}

func newTestContext(delivery *fakeDelivery, publisher *fakePublisher, headers amqp.Table) *RabbitContext {
	ctx := NewContext(delivery, headers, "sophia-resource.event.action.all", publisher.publish)
	ctx.DeliveryStruct = DeliveryStruct{DeliveryInterface: delivery, Body: []byte(`{"k":"v"}`)}
	return ctx
}

func TestSuccessAcks(t *testing.T) {
	delivery := &fakeDelivery{}
	ctx := newTestContext(delivery, &fakePublisher{}, nil)

	ctx.Success()

	if delivery.acks != 1 {
		t.Fatalf("acks = %d, want 1", delivery.acks)
	}
}

func TestRejectDeadLettersWithoutRequeue(t *testing.T) {
	delivery := &fakeDelivery{}
	ctx := newTestContext(delivery, &fakePublisher{}, nil)

	ctx.Reject()

	if len(delivery.rejects) != 1 || delivery.rejects[0] != false {
		t.Fatalf("rejects = %v, want one Reject(requeue=false)", delivery.rejects)
	}
}

func TestFailRepublishesToRetryQueueWhileUnderLimit(t *testing.T) {
	tests := []struct {
		name        string
		headers     amqp.Table
		wantRetries int32
	}{
		{"first failure, nil headers", nil, 1},
		{"second failure", amqp.Table{"x-event-people-retries": int32(1)}, 2},
		{"int64 header from broker", amqp.Table{"x-event-people-retries": int64(2)}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delivery := &fakeDelivery{}
			publisher := &fakePublisher{}
			ctx := newTestContext(delivery, publisher, tt.headers)

			ctx.Fail()

			if len(publisher.routingKeys) != 1 || publisher.routingKeys[0] != "sophia-resource.event.action.all.retry" {
				t.Fatalf("published to %v, want retry queue", publisher.routingKeys)
			}
			if got := publisher.headers[0]["x-event-people-retries"]; got != tt.wantRetries {
				t.Errorf("retries header = %v, want %d", got, tt.wantRetries)
			}
			if string(publisher.bodies[0]) != `{"k":"v"}` {
				t.Errorf("body = %s, want original body", publisher.bodies[0])
			}
			if delivery.acks != 1 {
				t.Errorf("acks = %d, want 1 (original message acked after republish)", delivery.acks)
			}
			if len(delivery.nacks) != 0 {
				t.Errorf("nacks = %v, want none", delivery.nacks)
			}
		})
	}
}

func TestFailDeadLettersWhenRetriesExhausted(t *testing.T) {
	delivery := &fakeDelivery{}
	publisher := &fakePublisher{}
	headers := amqp.Table{"x-event-people-retries": int32(3)}
	ctx := newTestContext(delivery, publisher, headers)

	ctx.Fail()

	if len(publisher.routingKeys) != 0 {
		t.Errorf("published %v, want no republish", publisher.routingKeys)
	}
	if len(delivery.nacks) != 1 || delivery.nacks[0] != false {
		t.Fatalf("nacks = %v, want one Nack(requeue=false)", delivery.nacks)
	}
	if delivery.acks != 0 {
		t.Errorf("acks = %d, want 0", delivery.acks)
	}
}

func TestFailDeadLettersWhenRepublishFails(t *testing.T) {
	delivery := &fakeDelivery{}
	publisher := &fakePublisher{err: errors.New("channel closed")}
	ctx := newTestContext(delivery, publisher, nil)

	ctx.Fail()

	if len(delivery.nacks) != 1 || delivery.nacks[0] != false {
		t.Fatalf("nacks = %v, want one Nack(requeue=false) fallback", delivery.nacks)
	}
	if delivery.acks != 0 {
		t.Errorf("acks = %d, want 0", delivery.acks)
	}
}

func TestFailWithoutPublisherFallsBackToDeadLetter(t *testing.T) {
	delivery := &fakeDelivery{}
	ctx := NewContext(delivery, nil, "queue", nil)
	ctx.DeliveryStruct = DeliveryStruct{DeliveryInterface: delivery, Body: []byte(`{}`)}

	ctx.Fail()

	if len(delivery.nacks) != 1 || delivery.nacks[0] != false {
		t.Fatalf("nacks = %v, want one Nack(requeue=false)", delivery.nacks)
	}
}
