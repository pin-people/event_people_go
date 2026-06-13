package EventPeople

import "os"

// RetryConfig holds the global retry configuration defaults.
type RetryConfig struct {
	MaxAttempts   int
	InitialDelay  int
	DelayStrategy string
	DLQName       string
}

type configStruct struct {
	Broker        AbstractBaseBroker
	MaxAttempts   int
	InitialDelay  int
	DelayStrategy string
	DLQName       string
}

var Config = &configStruct{
	MaxAttempts:   3,
	InitialDelay:  1000,
	DelayStrategy: "exponential",
}

// Configure sets global retry defaults in code. Connection attributes
// (appName, url, vhost, topic) are always read from environment variables.
// Options: MaxAttempts, InitialDelay, DelayStrategy, DLQName.
func (config *configStruct) Configure(options RetryConfig) {
	if options.MaxAttempts != 0 {
		config.MaxAttempts = options.MaxAttempts
	}
	if options.InitialDelay != 0 {
		config.InitialDelay = options.InitialDelay
	}
	if options.DelayStrategy != "" {
		config.DelayStrategy = options.DelayStrategy
	}
	if options.DLQName != "" {
		config.DLQName = options.DLQName
	}
}

// GetRetryConfig returns the active global retry configuration.
func (config *configStruct) GetRetryConfig() RetryConfig {
	dlqName := config.DLQName
	if dlqName == "" {
		dlqName = os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME") + "_dlq"
	}
	return RetryConfig{
		MaxAttempts:   config.MaxAttempts,
		InitialDelay:  config.InitialDelay,
		DelayStrategy: config.DelayStrategy,
		DLQName:       dlqName,
	}
}

func (config *configStruct) Init() {
	Config.Broker = new(RabbitBroker)
	Config.Broker.Init()
}

func (config *configStruct) CloseConnection() {
	Config.Broker.CloseConnection()
}
