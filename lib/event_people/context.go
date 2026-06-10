package EventPeople

import amqp "github.com/rabbitmq/amqp091-go"

type ContextInterface interface {
	Initialize(DeliveryInterface)
	Success()
	Fail()
	Reject()
}

type DeliveryInterface interface {
	Ack(bool) error
	Nack(bool, bool) error
	Reject(bool) error
}

type DeliveryStruct struct {
	DeliveryInterface
	DeliveryTag uint64
	Body        []byte
	Headers     amqp.Table
	QueueName   string
	Publish     PublishFunc
}
