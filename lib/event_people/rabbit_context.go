package EventPeople

import (
	"log"
	"os"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

// amqpPublisher is the narrow slice of *amqp.Channel that the context needs to
// publish messages (to the retry queue and to the DLQ). Declaring it as an
// interface keeps the publish paths unit-testable with a fake.
type amqpPublisher interface {
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
}

type RabbitContext struct {
	ContextInterface
	delivery       DeliveryInterface
	DeliveryStruct DeliveryStruct
	MaxRetries     int
	DLQName        string
	channel        amqpPublisher
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
				context.delivery.Nack(false, true)
				return
			}
		} else {
			log.Println("No channel available for retry publish; falling back to nack+requeue")
			context.delivery.Nack(false, true)
			return
		}
		context.delivery.Ack(false)
	} else {
		// Exhausted retries — route the message to the application-level DLQ
		// (a plain <app>_dlq queue) and ack the original delivery.
		if context.channel == nil {
			log.Println("No channel available for DLQ publish on exhaustion; falling back to nack")
			context.delivery.Nack(false, false)
			return
		}
		if err := context.publishToDLQ(); err != nil {
			log.Printf("Failed to publish to DLQ on retry exhaustion: %v", err)
			context.delivery.Nack(false, false)
			return
		}
		context.delivery.Ack(false)
	}
}

// Reject routes the message straight to the application-level DLQ (a plain
// <app>_dlq queue) and acks the original delivery. This follows the EventPeople
// spec (§C): the library routes to its own DLQ instead of relying on RabbitMQ's
// broker-side dead-letter-exchange. With no x-dead-letter-exchange argument on
// the main queue, library upgrades stay drop-in (no PRECONDITION_FAILED).
func (context RabbitContext) Reject() {
	if context.channel == nil {
		// No channel to publish with — fall back to a plain nack without requeue.
		log.Println("No channel available for DLQ publish on Reject; falling back to nack")
		context.delivery.Nack(false, false)
		return
	}
	if err := context.publishToDLQ(); err != nil {
		log.Printf("Failed to publish to DLQ on Reject: %v", err)
		context.delivery.Nack(false, false)
		return
	}
	context.delivery.Ack(false)
}

// publishToDLQ forwards the current message body to the application-level DLQ
// via the default exchange (routing key = DLQ name).
func (context RabbitContext) publishToDLQ() error {
	return context.channel.Publish(
		"",              // default exchange
		context.DLQName, // routing key = DLQ queue name
		false,
		false,
		amqp.Publishing{
			Body:        context.DeliveryStruct.Body,
			ContentType: context.DeliveryStruct.ContentType,
			Headers:     amqp.Table{"x-event-people-retries": int32(context.DeliveryStruct.RetryCount)},
		},
	)
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
