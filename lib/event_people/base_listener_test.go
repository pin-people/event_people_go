package EventPeople

import (
	"os"
	"testing"
)

func TestBaseListenerResolvedRetryConfigUsesGlobalDefaults(t *testing.T) {
	resetConfig()
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "testapp")

	bl := &BaseListener{}
	rc := bl.ResolvedRetryConfig()

	if rc.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3 from global default, got %d", rc.MaxAttempts)
	}
	if rc.InitialDelay != 1000 {
		t.Errorf("expected InitialDelay=1000 from global default, got %d", rc.InitialDelay)
	}
	if rc.DelayStrategy != "exponential" {
		t.Errorf("expected DelayStrategy=exponential from global default, got %s", rc.DelayStrategy)
	}
	if rc.DLQName != "testapp_dlq" {
		t.Errorf("expected DLQName=testapp_dlq from global default, got %s", rc.DLQName)
	}
}

func TestBaseListenerResolvedRetryConfigOverridesGlobal(t *testing.T) {
	resetConfig()
	defer resetConfig()
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "testapp")

	bl := &BaseListener{
		RetryConfig: ListenerRetryConfig{
			MaxAttempts:   10,
			InitialDelay:  500,
			DelayStrategy: "fixed",
			DLQName:       "listener_dlq",
		},
	}
	rc := bl.ResolvedRetryConfig()

	if rc.MaxAttempts != 10 {
		t.Errorf("expected MaxAttempts=10 from listener override, got %d", rc.MaxAttempts)
	}
	if rc.InitialDelay != 500 {
		t.Errorf("expected InitialDelay=500 from listener override, got %d", rc.InitialDelay)
	}
	if rc.DelayStrategy != "fixed" {
		t.Errorf("expected DelayStrategy=fixed from listener override, got %s", rc.DelayStrategy)
	}
	if rc.DLQName != "listener_dlq" {
		t.Errorf("expected DLQName=listener_dlq from listener override, got %s", rc.DLQName)
	}
}

func TestBaseListenerResolvedRetryConfigPartialOverride(t *testing.T) {
	resetConfig()
	defer resetConfig()
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "testapp")

	// Only override MaxAttempts; the rest should fall through to global defaults.
	bl := &BaseListener{
		RetryConfig: ListenerRetryConfig{
			MaxAttempts: 7,
		},
	}
	rc := bl.ResolvedRetryConfig()

	if rc.MaxAttempts != 7 {
		t.Errorf("expected MaxAttempts=7 from listener override, got %d", rc.MaxAttempts)
	}
	if rc.InitialDelay != 1000 {
		t.Errorf("expected InitialDelay=1000 from global default, got %d", rc.InitialDelay)
	}
	if rc.DelayStrategy != "exponential" {
		t.Errorf("expected DelayStrategy=exponential from global default, got %s", rc.DelayStrategy)
	}
}

func TestFixedEventNameThreeParts(t *testing.T) {
	name := FixedEventName("user.auth.created", "all")
	if name != "user.auth.created.all" {
		t.Errorf("expected user.auth.created.all, got %s", name)
	}
}

func TestFixedEventNameFourParts(t *testing.T) {
	name := FixedEventName("user.auth.created.other", "myapp")
	if name != "user.auth.created.myapp" {
		t.Errorf("expected user.auth.created.myapp, got %s", name)
	}
}
