package EventPeople

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// PublishFunc publishes a message body with headers to a routing key on
// the default exchange. Used to send failed messages to the retry queue.
type PublishFunc func(routingKey string, headers amqp.Table, body []byte) error

const retriesHeader = "x-event-people-retries"

type RabbitContext struct {
	ContextInterface
	delivery       DeliveryInterface
	DeliveryStruct DeliveryStruct
	headers        amqp.Table
	queueName      string
	publish        PublishFunc
}

func (context RabbitContext) Initialize(delivery DeliveryInterface) {
	context.delivery = delivery
}

func (context RabbitContext) Success() {
	context.delivery.Ack(false)
}

// Fail retries the message a bounded number of times by republishing it
// to the retry queue, then dead-letters it once retries are exhausted or
// republishing is not possible.
func (context RabbitContext) Fail() {
	if context.publish == nil || context.retryCount() >= MaxRetries() {
		context.deadLetter()
		return
	}

	headers := amqp.Table{}
	for key, value := range context.headers {
		headers[key] = value
	}
	headers[retriesHeader] = int32(context.retryCount() + 1)

	err := context.publish(RetryQueueName(context.queueName), headers, context.DeliveryStruct.Body)
	if err != nil {
		log.Printf("event_people: retry republish failed, dead-lettering: %v", err)
		context.deadLetter()
		return
	}
	context.delivery.Ack(false)
}

func (context RabbitContext) Reject() {
	context.delivery.Reject(false)
}

// deadLetter routes the message to the DLQ via the queue's
// x-dead-letter-exchange (Nack without requeue).
func (context RabbitContext) deadLetter() {
	context.delivery.Nack(false, false)
}

func (context RabbitContext) retryCount() int {
	switch value := context.headers[retriesHeader].(type) {
	case int32:
		return int(value)
	case int64:
		return int(value)
	case int:
		return value
	default:
		return 0
	}
}

func NewContext(delivery DeliveryInterface, headers amqp.Table, queueName string, publish PublishFunc) *RabbitContext {
	return &RabbitContext{
		delivery:  delivery,
		headers:   headers,
		queueName: queueName,
		publish:   publish,
	}
}
