package EventPeople

import (
	"os"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

// fakeChannel records queue declarations so we can assert the topology without a broker.
type fakeChannel struct {
	declaredQueues map[string]amqp.Table
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
func (f *fakeChannel) ExchangeDeclarePassive(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	return nil
}
func (f *fakeChannel) Qos(prefetchCount, prefetchSize int, global bool) error { return nil }
func (f *fakeChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return nil, nil
}

// The main queue must be declared WITHOUT x-dead-letter-exchange (so upgrades over
// legacy queues never PRECONDITION_FAIL), and the DLQ must be a plain queue.
func TestCreateQueueAndBind_NoDLXArg_PlainDLQ(t *testing.T) {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "sophia")
	os.Setenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME", "pinpeople")
	defer os.Unsetenv("RABBIT_EVENT_PEOPLE_APP_NAME")
	defer os.Unsetenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME")

	fc := &fakeChannel{}
	q := &Queue{channel: fc}
	if err := q.createQueueAndBind("resource.core.created"); err != nil {
		t.Fatalf("createQueueAndBind: %v", err)
	}

	const mainQueue = "sophia-resource.core.created.all"
	args, ok := fc.declaredQueues[mainQueue]
	if !ok {
		t.Fatalf("main queue %q not declared; declared: %v", mainQueue, fc.declaredQueues)
	}
	if _, bad := args["x-dead-letter-exchange"]; bad {
		t.Errorf("main queue must not declare x-dead-letter-exchange, got %v", args)
	}
	dlqArgs, ok := fc.declaredQueues["sophia_dlq"]
	if !ok {
		t.Fatalf("DLQ sophia_dlq not declared; declared: %v", fc.declaredQueues)
	}
	if dlqArgs != nil {
		t.Errorf("DLQ must be a plain queue (nil args), got %v", dlqArgs)
	}
	if _, ok := fc.declaredQueues[mainQueue+"_retry"]; !ok {
		t.Errorf("retry queue should still be declared")
	}
}

func TestQueueNameByRoutingKeyThreeParts(t *testing.T) {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "myapp")
	q := &Queue{}
	name := q.queueNameByRoutingKey("user.auth.created")
	expected := "myapp-user.auth.created.all"
	if name != expected {
		t.Errorf("expected %s, got %s", expected, name)
	}
}

func TestQueueNameByRoutingKeyFourPartsAll(t *testing.T) {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "myapp")
	q := &Queue{}
	name := q.queueNameByRoutingKey("user.auth.created.all")
	expected := "myapp-user.auth.created.all"
	if name != expected {
		t.Errorf("expected %s, got %s", expected, name)
	}
}

func TestQueueNameByRoutingKeyFourPartsAppName(t *testing.T) {
	// DEV-GO-002 fix: targeted routing key must also resolve to the .all queue name.
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "myapp")
	q := &Queue{}
	name := q.queueNameByRoutingKey("user.auth.created.myapp")
	expected := "myapp-user.auth.created.all"
	if name != expected {
		t.Errorf("DEV-GO-002: expected destination normalized to 'all', got %s (expected %s)", name, expected)
	}
}

func TestQueueNameByRoutingKeyFourPartsOtherDestination(t *testing.T) {
	// Any 4-part routing key with arbitrary destination must still produce .all queue.
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "svc")
	q := &Queue{}
	name := q.queueNameByRoutingKey("order.payments.paid.billing_service")
	expected := "svc-order.payments.paid.all"
	if name != expected {
		t.Errorf("expected destination normalized to 'all', got %s", name)
	}
}
