package EventPeople

import (
	"os"
	"testing"
)

func TestEventRetryCountDefaultsToZero(t *testing.T) {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "testapp")
	event := NewEvent("user.auth.created", map[string]string{"key": "value"})
	if event.RetryCount != 0 {
		t.Errorf("expected RetryCount=0, got %d", event.RetryCount)
	}
}

func TestEventIncrementRetryCount(t *testing.T) {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "testapp")
	event := NewEvent("user.auth.created", map[string]string{"key": "value"})
	event.IncrementRetryCount()
	if event.RetryCount != 1 {
		t.Errorf("expected RetryCount=1 after increment, got %d", event.RetryCount)
	}
	event.IncrementRetryCount()
	if event.RetryCount != 2 {
		t.Errorf("expected RetryCount=2 after second increment, got %d", event.RetryCount)
	}
}

func TestEventFixNameThreeParts(t *testing.T) {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "myapp")
	event := NewEvent("user.auth.created", nil)
	if event.Name != "user.auth.created.all" {
		t.Errorf("expected name to be appended with .all, got %s", event.Name)
	}
}

func TestEventFixNameFourParts(t *testing.T) {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "myapp")
	event := NewEvent("user.auth.created.payment", nil)
	if event.Name != "user.auth.created.payment" {
		t.Errorf("expected name unchanged at 4 parts, got %s", event.Name)
	}
}

func TestEventNameNoAppNamePrefix(t *testing.T) {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "myapp")
	event := NewEvent("user.auth.created.all", nil)
	// Routing key must NOT include appName prefix.
	if event.Name == "myapp-user.auth.created.all" {
		t.Errorf("event.Name must not have appName prefix — got %s", event.Name)
	}
	if event.Name != "user.auth.created.all" {
		t.Errorf("expected name=user.auth.created.all, got %s", event.Name)
	}
}

func TestEventHeaders(t *testing.T) {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "testapp")
	event := NewEvent("user.auth.created.all", nil)
	if event.Headers.AppName != "testapp" {
		t.Errorf("expected Headers.AppName=testapp, got %s", event.Headers.AppName)
	}
	if event.Headers.Resource != "user" {
		t.Errorf("expected Headers.Resource=user, got %s", event.Headers.Resource)
	}
	if event.Headers.Origin != "auth" {
		t.Errorf("expected Headers.Origin=auth, got %s", event.Headers.Origin)
	}
	if event.Headers.Action != "created" {
		t.Errorf("expected Headers.Action=created, got %s", event.Headers.Action)
	}
	if event.Headers.Destination != "all" {
		t.Errorf("expected Headers.Destination=all, got %s", event.Headers.Destination)
	}
}
