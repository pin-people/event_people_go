package EventPeople

import amqp "github.com/rabbitmq/amqp091-go"

type AbstractBaseBroker interface {
	Init() error
	GetConnection() amqp.Connection
	GetConsumers() int
	Subscribe(eventName string) error
	Consume(eventName string, callback Callback)
	Produce(events Event) error
	CloseConnection()
}

type BaseBroker struct {
	AbstractBaseBroker
}
