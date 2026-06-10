package EventPeople

import (
	"os"
	"strconv"
)

const (
	defaultMaxRetries = 3
	defaultRetryTTLMs = 30000
)

// MaxRetries returns how many times a failed message is retried before
// being dead-lettered. Configured via RABBIT_EVENT_PEOPLE_MAX_RETRIES.
func MaxRetries() int {
	value, err := strconv.Atoi(os.Getenv("RABBIT_EVENT_PEOPLE_MAX_RETRIES"))
	if err != nil || value <= 0 {
		return defaultMaxRetries
	}
	return value
}

// RetryTTLMs returns how long a message waits in the retry queue before
// being redelivered. Configured via RABBIT_EVENT_PEOPLE_RETRY_TTL_MS.
func RetryTTLMs() int64 {
	value, err := strconv.ParseInt(os.Getenv("RABBIT_EVENT_PEOPLE_RETRY_TTL_MS"), 10, 64)
	if err != nil || value <= 0 {
		return defaultRetryTTLMs
	}
	return value
}

// DeadLetterExchangeName is the per-app fanout exchange that receives
// rejected and retry-exhausted messages.
func DeadLetterExchangeName() string {
	return os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME") + ".dlx"
}

// DeadLetterQueueName is the per-app parking queue bound to the DLX.
func DeadLetterQueueName() string {
	return os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME") + ".dlq"
}

// RetryQueueName is the per-queue wait queue used for bounded retries.
func RetryQueueName(queueName string) string {
	return queueName + ".retry"
}
