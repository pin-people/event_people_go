package EventPeople

import (
	"log"
	"os"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitContext struct {
	ContextInterface
	delivery       DeliveryInterface
	DeliveryStruct DeliveryStruct
	MaxRetries     int
	DLQName        string
	channel        *amqp.Channel
	initialDelay   int // resolved once at construction time from RABBIT_EVENT_PEOPLE_RETRY_TTL_MS
}

// IsLastRetry returns true when the current retry count is at or beyond the last allowed retry.
func (context *RabbitContext) IsLastRetry() bool {
	return context.DeliveryStruct.RetryCount >= context.MaxRetries-1
}

func (context RabbitContext) Initialize(delivery DeliveryInterface) {
	context.delivery = delivery
}

func (context RabbitContext) Success() {
	context.delivery.Ack(false)
}

// Fail publishes the message to the retry queue if retries remain, otherwise Nacks without requeue.
func (context *RabbitContext) Fail() {
	retryCount := context.DeliveryStruct.RetryCount
	maxRetries := context.MaxRetries
	queueName := context.DeliveryStruct.QueueName
	delayStrategy := context.DeliveryStruct.DelayStrategy

	rm := NewRetryManagerWithDelay(maxRetries, delayStrategy, context.initialDelay)

	if rm.ShouldRetry(retryCount) {
		retryQueueName := queueName + "_retry"
		delay := rm.GetNextDelay(retryCount)

		if context.channel != nil {
			err := context.channel.Publish(
				"",              // default exchange
				retryQueueName, // routing key = retry queue name
				false,
				false,
				amqp.Publishing{
					Headers:     amqp.Table{"x-event-people-retries": int32(retryCount + 1)},
					Expiration:  strconv.Itoa(delay),
					Body:        context.DeliveryStruct.Body,
					ContentType: context.DeliveryStruct.ContentType,
				},
			)
			if err != nil {
				log.Printf("Failed to publish to retry queue: %v", err)
				context.delivery.Nack(false, false)
				return
			}
		} else {
			log.Println("No channel available for retry publish; falling back to nack without requeue (→ DLQ)")
			context.delivery.Nack(false, false)
			return
		}
		context.delivery.Ack(false)
	} else {
		// Exhausted retries — send to DLQ via nack without requeue (DLX will route to DLQ)
		context.delivery.Nack(false, false)
	}
}

// Reject nacks without requeue, triggering DLX → DLQ routing.
func (context RabbitContext) Reject() {
	context.delivery.Nack(false, false)
}

// NewContext creates a RabbitContext with delivery only (no retry config).
// Used for backward-compatibility paths.
func NewContext(delivery DeliveryInterface) *RabbitContext {
	appName := os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME")
	context := &RabbitContext{
		delivery:     delivery,
		MaxRetries:   Config.MaxAttempts(),
		DLQName:      appName + "_dlq",
		initialDelay: resolveInitialDelay(),
	}
	return context
}

// NewContextWithRetry creates a RabbitContext with full retry configuration.
func NewContextWithRetry(delivery DeliveryInterface, channel *amqp.Channel, queueName string, maxRetries int, delayStrategy string, retryCount int) *RabbitContext {
	appName := os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME")
	ctx := &RabbitContext{
		delivery:     delivery,
		MaxRetries:   maxRetries,
		DLQName:      appName + "_dlq",
		channel:      channel,
		initialDelay: resolveInitialDelay(),
	}
	ctx.DeliveryStruct.MaxRetries = maxRetries
	ctx.DeliveryStruct.DelayStrategy = delayStrategy
	ctx.DeliveryStruct.QueueName = queueName
	ctx.DeliveryStruct.RetryCount = retryCount
	return ctx
}
