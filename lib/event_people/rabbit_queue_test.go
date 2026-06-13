package EventPeople

import (
	"os"
	"testing"
)

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
