package EventPeople

import (
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

var mutex sync.Mutex

type RabbitContext struct {
	ContextInterface
	delivery *amqp.Delivery
}

func (context *RabbitContext) Initialize(delivery *amqp.Delivery) {
	context.delivery = delivery
}

func (context *RabbitContext) Success() {
	context.delivery.Ack(false)
	mutex.Unlock()
}

func (context *RabbitContext) Fail() {
	context.delivery.Nack(false, true)
	mutex.Unlock()
}

func (context *RabbitContext) Reject() {
	context.delivery.Reject(false)
	mutex.Unlock()
}

func NewContext(delivery *amqp.Delivery) *RabbitContext {
	context := new(RabbitContext)
	context.Initialize(delivery)
	mutex.Lock()
	return context
}
