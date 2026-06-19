package EventPeople

import (
	"context"
	"fmt"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// retryPublisher is the minimal interface needed by RabbitContext.Fail() to
// publish a message to the retry queue. *amqp.Channel satisfies this interface.
type retryPublisher interface {
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
}

// RabbitContext is the RabbitMQ-specific implementation of ContextInterface.
// It is injected into every listener callback and manages ack/nack/retry/DLQ routing.
type RabbitContext struct {
	delivery       DeliveryInterface
	DeliveryStruct DeliveryStruct
	// MaxRetries is the total retry attempts configured for this listener.
	MaxRetries int
	// IsLastRetry is true when this is the final retry attempt.
	IsLastRetry bool
	// DLQName is the dead-letter queue name for this listener.
	DLQName string
	// InitialDelay is the base retry delay in milliseconds resolved for this listener.
	InitialDelay int
	// DelayStrategy is the retry delay strategy resolved for this listener ("fixed" or "exponential").
	DelayStrategy string
	// retryQueueName is the name of the per-queue retry queue.
	retryQueueName string
	// retryCount is the current retry attempt count from the event.
	retryCount int
	// amqpChannel is used to publish to the retry queue.
	amqpChannel retryPublisher
}

// NewContext creates a RabbitContext with default (no-retry) configuration.
// Use NewContextWithRetry to attach retry parameters.
func NewContext(delivery DeliveryInterface) *RabbitContext {
	ctx := &RabbitContext{delivery: delivery}
	return ctx
}

// NewContextWithRetry creates a fully configured RabbitContext for retry-aware dispatch.
func NewContextWithRetry(
	delivery DeliveryInterface,
	amqpChannel retryPublisher,
	retryQueueName string,
	retryCount int,
	maxRetries int,
	dlqName string,
	initialDelay int,
	delayStrategy string,
) *RabbitContext {
	return &RabbitContext{
		delivery:       delivery,
		amqpChannel:    amqpChannel,
		retryQueueName: retryQueueName,
		retryCount:     retryCount,
		MaxRetries:     maxRetries,
		IsLastRetry:    retryCount >= maxRetries-1,
		DLQName:        dlqName,
		InitialDelay:   initialDelay,
		DelayStrategy:  delayStrategy,
	}
}

func (ctx *RabbitContext) Initialize(delivery DeliveryInterface) {
	ctx.delivery = delivery
}

// GetMaxRetries satisfies ContextInterface.
func (ctx *RabbitContext) GetMaxRetries() int {
	return ctx.MaxRetries
}

// GetIsLastRetry satisfies ContextInterface.
func (ctx *RabbitContext) GetIsLastRetry() bool {
	return ctx.IsLastRetry
}

// Success acknowledges successful processing (ack). Removes message from queue.
func (ctx *RabbitContext) Success() {
	ctx.delivery.Ack(false)
}

// Fail indicates failure. Publishes the message to the retry queue with a backoff
// delay when retries remain, then acks the original delivery. When retries are
// exhausted, or when publishing to the retry queue itself fails (broker unavailable,
// channel closed), MUST nack(requeue=false) so the DLX routes the message to the DLQ.
// Nacking with requeue=true on publish failure causes an infinite redelivery loop
// because x-event-people-retries is never incremented.
func (ctx *RabbitContext) Fail() {
	// No retry channel — legacy or no-retry path: nack without requeue → DLQ via DLX.
	if ctx.amqpChannel == nil || ctx.retryQueueName == "" {
		ctx.delivery.Nack(false, false)
		return
	}

	rm := NewRetryManager(ctx.MaxRetries, ctx.InitialDelay, ctx.DelayStrategy)
	rm.CurrentAttempt = ctx.retryCount

	if !rm.ShouldRetry() {
		// Retries exhausted: publish the message to the application-level DLQ + ack.
		if err := ctx.publishToDLQ(); err != nil {
			fmt.Printf("EventPeople: failed to publish to DLQ %s on retry exhaustion: %v — nacking\n", ctx.DLQName, err)
			ctx.delivery.Nack(false, false)
			return
		}
		ctx.delivery.Ack(false)
		return
	}

	delay := rm.GetNextDelay()
	nextRetryCount := ctx.retryCount + 1

	headers := amqp.Table{
		"x-event-people-retries": int64(nextRetryCount),
	}

	body := ctx.DeliveryStruct.Body
	if body == nil {
		// Fallback to delivery body if DeliveryStruct wasn't populated.
		if ds, ok := ctx.delivery.(interface{ GetBody() []byte }); ok {
			body = ds.GetBody()
		}
	}

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "text/plain",
		Body:         body,
		Headers:      headers,
		Expiration:   strconv.Itoa(delay),
	}

	pubCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := ctx.amqpChannel.PublishWithContext(
		pubCtx,
		"",                 // default exchange
		ctx.retryQueueName, // routing key = queue name
		false,
		false,
		msg,
	)
	if err != nil {
		// Publish to retry queue failed: nack(requeue=false) → DLX → DLQ.
		// Never nack(requeue=true) — that causes an infinite redelivery loop.
		fmt.Printf("EventPeople: failed to publish to retry queue %s: %v — nacking to DLQ\n", ctx.retryQueueName, err)
		ctx.delivery.Nack(false, false)
		return
	}

	// Successfully published to retry queue: ack the original delivery.
	ctx.delivery.Ack(false)
}

// Reject rejects the event. Routes directly to DLQ without retrying (nack, requeue=false).
func (ctx *RabbitContext) Reject() {
	if ctx.amqpChannel == nil || ctx.DLQName == "" {
		ctx.delivery.Nack(false, false)
		return
	}
	if err := ctx.publishToDLQ(); err != nil {
		fmt.Printf("EventPeople: failed to publish to DLQ %s on Reject: %v — nacking\n", ctx.DLQName, err)
		ctx.delivery.Nack(false, false)
		return
	}
	ctx.delivery.Ack(false)
}

// publishToDLQ forwards the current message body to the application-level DLQ via
// the default exchange (routing key = DLQ name), so failed messages are dead-lettered
// without relying on a broker dead-letter-exchange.
func (ctx *RabbitContext) publishToDLQ() error {
	pubCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// The original message may not have its ContentType populated upstream.
	// An empty ContentType produces an unparseable DLQ message, so fall back to
	// the same value used by every other publish path in this implementation.
	contentType := ctx.DeliveryStruct.ContentType
	if contentType == "" {
		contentType = "text/plain"
	}

	return ctx.amqpChannel.PublishWithContext(
		pubCtx,
		"",          // default exchange
		ctx.DLQName, // routing key = DLQ queue name
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			ContentType:  contentType,
			Body:         ctx.DeliveryStruct.Body,
			Headers:      amqp.Table{"x-event-people-retries": int64(ctx.retryCount)},
		},
	)
}
