package EventPeople

import (
	"os"
	"strconv"
)

type RetryConfig struct {
	MaxAttempts   int
	DelayStrategy string
	DLQName       string
}

type configStruct struct {
	Broker AbstractBaseBroker
}

var Config = new(configStruct)

func (config *configStruct) Init() {
	Config.Broker = new(RabbitBroker)
	Config.Broker.Init()
}

func (config *configStruct) CloseConnection() {
	Config.Broker.CloseConnection()
}

// MaxAttempts returns the maximum number of retry attempts from
// RABBIT_EVENT_PEOPLE_MAX_RETRIES (default 3).
func (config *configStruct) MaxAttempts() int {
	if v := os.Getenv("RABBIT_EVENT_PEOPLE_MAX_RETRIES"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			return parsed
		}
	}
	return 3
}

// DelayStrategy returns the delay strategy (default "exponential").
func (config *configStruct) DelayStrategy() string {
	return "exponential"
}

// DLQName returns the dead-letter queue name (default "{APP_NAME}_dlq").
func (config *configStruct) DLQName() string {
	appName := os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME")
	return appName + "_dlq"
}

// GetRetryConfig returns the full retry configuration using the current defaults.
func (config *configStruct) GetRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:   config.MaxAttempts(),
		DelayStrategy: config.DelayStrategy(),
		DLQName:       config.DLQName(),
	}
}
