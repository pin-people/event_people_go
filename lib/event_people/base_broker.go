package EventPeople

import amqp "github.com/rabbitmq/amqp091-go"

type AbstractBaseBroker interface {
	Init()
	GetConnection() amqp.Connection
	GetConsumers() int
	Subscribe(eventName string)
	Consume(eventName string) *DeliveryStruct
	Produce(events Event)
	CloseConnection()
}

type BaseBroker struct {
	AbstractBaseBroker
}
