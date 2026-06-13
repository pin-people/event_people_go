package EventPeople

import (
	"os"
	"testing"
)

func resetConfig() {
	Config.MaxAttempts = 3
	Config.InitialDelay = 1000
	Config.DelayStrategy = "exponential"
	Config.DLQName = ""
}

func TestConfigDefaults(t *testing.T) {
	resetConfig()

	if Config.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", Config.MaxAttempts)
	}
	if Config.InitialDelay != 1000 {
		t.Errorf("expected InitialDelay=1000, got %d", Config.InitialDelay)
	}
	if Config.DelayStrategy != "exponential" {
		t.Errorf("expected DelayStrategy=exponential, got %s", Config.DelayStrategy)
	}
}

func TestConfigConfigure(t *testing.T) {
	resetConfig()
	defer resetConfig()

	Config.Configure(RetryConfig{
		MaxAttempts:   5,
		InitialDelay:  2000,
		DelayStrategy: "fixed",
		DLQName:       "my_dlq",
	})

	if Config.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts=5, got %d", Config.MaxAttempts)
	}
	if Config.InitialDelay != 2000 {
		t.Errorf("expected InitialDelay=2000, got %d", Config.InitialDelay)
	}
	if Config.DelayStrategy != "fixed" {
		t.Errorf("expected DelayStrategy=fixed, got %s", Config.DelayStrategy)
	}
	if Config.DLQName != "my_dlq" {
		t.Errorf("expected DLQName=my_dlq, got %s", Config.DLQName)
	}
}

func TestConfigConfigurePartial(t *testing.T) {
	resetConfig()
	defer resetConfig()

	// Only override MaxAttempts; others should retain defaults.
	Config.Configure(RetryConfig{MaxAttempts: 7})

	if Config.MaxAttempts != 7 {
		t.Errorf("expected MaxAttempts=7, got %d", Config.MaxAttempts)
	}
	if Config.InitialDelay != 1000 {
		t.Errorf("expected InitialDelay=1000 (unchanged), got %d", Config.InitialDelay)
	}
	if Config.DelayStrategy != "exponential" {
		t.Errorf("expected DelayStrategy=exponential (unchanged), got %s", Config.DelayStrategy)
	}
}

func TestConfigGetRetryConfigDefaultDLQ(t *testing.T) {
	resetConfig()
	defer resetConfig()
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "testapp")

	rc := Config.GetRetryConfig()

	if rc.DLQName != "testapp_dlq" {
		t.Errorf("expected DLQName=testapp_dlq, got %s", rc.DLQName)
	}
	if rc.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", rc.MaxAttempts)
	}
	if rc.InitialDelay != 1000 {
		t.Errorf("expected InitialDelay=1000, got %d", rc.InitialDelay)
	}
	if rc.DelayStrategy != "exponential" {
		t.Errorf("expected DelayStrategy=exponential, got %s", rc.DelayStrategy)
	}
}

func TestConfigGetRetryConfigExplicitDLQ(t *testing.T) {
	resetConfig()
	defer resetConfig()

	Config.Configure(RetryConfig{DLQName: "custom_dlq"})

	rc := Config.GetRetryConfig()
	if rc.DLQName != "custom_dlq" {
		t.Errorf("expected DLQName=custom_dlq, got %s", rc.DLQName)
	}
}
