package EventPeople

import (
    amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitContext struct {
    ContextInterface
    delivery *amqp.Delivery
}

func (context *RabbitContext) Initialize(delivery *amqp.Delivery) {
    context.delivery = delivery
}

func (context *RabbitContext) Success() {
    context.delivery.Ack(false)
}

func (context *RabbitContext) Fail() {
    context.delivery.Nack(false, true)
}

func (context *RabbitContext) Reject() {
    context.delivery.Reject(false)
}

func NewContext(delivery *amqp.Delivery) *RabbitContext {
    context := new(RabbitContext)
    context.Initialize(delivery)
    return context
}
