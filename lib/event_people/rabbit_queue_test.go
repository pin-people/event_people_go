package EventPeople

import (
	"os"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

// fakeChannel records declarations so we can assert the queue topology without
// a live broker.
type fakeChannel struct {
	declaredQueues    map[string]amqp.Table // queue name -> declare args
	declaredExchanges []string              // active (non-passive) exchange declares
}

func (f *fakeChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	if f.declaredQueues == nil {
		f.declaredQueues = map[string]amqp.Table{}
	}
	f.declaredQueues[name] = args
	return amqp.Queue{Name: name}, nil
}
func (f *fakeChannel) QueueInspect(name string) (amqp.Queue, error) { return amqp.Queue{Name: name}, nil }
func (f *fakeChannel) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	return nil
}
func (f *fakeChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	f.declaredExchanges = append(f.declaredExchanges, name)
	return nil
}
func (f *fakeChannel) ExchangeDeclarePassive(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	return nil
}
func (f *fakeChannel) Qos(prefetchCount, prefetchSize int, global bool) error { return nil }
func (f *fakeChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return nil, nil
}

// The main queue must be declared WITHOUT x-dead-letter-exchange so that
// upgrading over legacy (arg-less) queues never triggers PRECONDITION_FAILED,
// and the DLQ must be a plain queue (no broker dead-lettering topology).
func TestCreateQueueAndBind_NoDLXArg_PlainDLQ(t *testing.T) {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "sophia")
	os.Setenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME", "pinpeople")
	defer os.Unsetenv("RABBIT_EVENT_PEOPLE_APP_NAME")
	defer os.Unsetenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME")

	fc := &fakeChannel{}
	q := &Queue{channel: fc}

	if err := q.createQueueAndBind("resource.core.created"); err != nil {
		t.Fatalf("createQueueAndBind returned error: %v", err)
	}

	const mainQueue = "sophia-resource.core.created.all"
	args, ok := fc.declaredQueues[mainQueue]
	if !ok {
		t.Fatalf("main queue %q was not declared; declared: %v", mainQueue, fc.declaredQueues)
	}
	if _, bad := args["x-dead-letter-exchange"]; bad {
		t.Errorf("main queue must not declare x-dead-letter-exchange, got args %v", args)
	}

	dlqArgs, ok := fc.declaredQueues["sophia_dlq"]
	if !ok {
		t.Fatalf("DLQ sophia_dlq was not declared; declared: %v", fc.declaredQueues)
	}
	if dlqArgs != nil {
		t.Errorf("DLQ must be a plain queue (nil args), got %v", dlqArgs)
	}

	if len(fc.declaredExchanges) != 0 {
		t.Errorf("expected no active exchange declares (DLX fanout removed), got %v", fc.declaredExchanges)
	}

	// Retry queue is still declared (TTL backoff path is unchanged).
	if _, ok := fc.declaredQueues[mainQueue+"_retry"]; !ok {
		t.Errorf("retry queue %q_retry should still be declared", mainQueue)
	}
}
