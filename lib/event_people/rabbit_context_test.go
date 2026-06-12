package EventPeople

import (
	"errors"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

// fakeDelivery records the last Ack/Nack call made by RabbitContext.
type fakeDelivery struct {
	ackCalled    bool
	nackCalled   bool
	nackMultiple bool
	nackRequeue  bool
}

func (f *fakeDelivery) Ack(multiple bool) error {
	f.ackCalled = true
	return nil
}

func (f *fakeDelivery) Nack(multiple, requeue bool) error {
	f.nackCalled = true
	f.nackMultiple = multiple
	f.nackRequeue = requeue
	return nil
}

func (f *fakeDelivery) Reject(requeue bool) error { return nil }

func publishErr(_, _ string, _, _ bool, _ amqp.Publishing) error {
	return errors.New("broker unavailable")
}

func publishOK(_, _ string, _, _ bool, _ amqp.Publishing) error { return nil }

func baseCtx(d *fakeDelivery, retryCount, maxRetries int) *RabbitContext {
	ctx := &RabbitContext{
		delivery:     d,
		MaxRetries:   maxRetries,
		initialDelay: 1000,
	}
	ctx.DeliveryStruct.RetryCount = retryCount
	ctx.DeliveryStruct.DelayStrategy = "exponential"
	ctx.DeliveryStruct.QueueName = "test_queue"
	ctx.DeliveryStruct.MaxRetries = maxRetries
	return ctx
}

// TestFail_NilPublishFn_NacksWithoutRequeue: no channel configured → DLQ via nack(false,false).
func TestFail_NilPublishFn_NacksWithoutRequeue(t *testing.T) {
	d := &fakeDelivery{}
	ctx := baseCtx(d, 0, 3) // retries remain, but publishFn is nil
	ctx.Fail()
	if !d.nackCalled || d.nackMultiple || d.nackRequeue {
		t.Errorf("expected Nack(false, false); got nackCalled=%v multiple=%v requeue=%v",
			d.nackCalled, d.nackMultiple, d.nackRequeue)
	}
	if d.ackCalled {
		t.Error("Ack must not be called when publishFn is nil")
	}
}

// TestFail_PublishError_NacksWithoutRequeue: broker error → DLQ via nack(false,false), not requeue.
func TestFail_PublishError_NacksWithoutRequeue(t *testing.T) {
	d := &fakeDelivery{}
	ctx := baseCtx(d, 0, 3)
	ctx.publishFn = publishErr
	ctx.Fail()
	if !d.nackCalled || d.nackMultiple || d.nackRequeue {
		t.Errorf("expected Nack(false, false); got nackCalled=%v multiple=%v requeue=%v",
			d.nackCalled, d.nackMultiple, d.nackRequeue)
	}
	if d.ackCalled {
		t.Error("Ack must not be called on publish failure")
	}
}

// TestFail_RetriesExhausted_NacksWithoutRequeue: no retries left → DLQ via nack(false,false).
func TestFail_RetriesExhausted_NacksWithoutRequeue(t *testing.T) {
	d := &fakeDelivery{}
	ctx := baseCtx(d, 3, 3) // retryCount == maxRetries
	ctx.publishFn = publishOK
	ctx.Fail()
	if !d.nackCalled || d.nackMultiple || d.nackRequeue {
		t.Errorf("expected Nack(false, false); got nackCalled=%v multiple=%v requeue=%v",
			d.nackCalled, d.nackMultiple, d.nackRequeue)
	}
}

// TestFail_PublishSuccess_Acks: publish succeeds → ack the original delivery.
func TestFail_PublishSuccess_Acks(t *testing.T) {
	d := &fakeDelivery{}
	ctx := baseCtx(d, 0, 3)
	ctx.publishFn = publishOK
	ctx.Fail()
	if !d.ackCalled {
		t.Error("expected Ack after successful retry publish")
	}
	if d.nackCalled {
		t.Error("Nack must not be called on successful retry publish")
	}
}
