package EventPeople

import amqp "github.com/rabbitmq/amqp091-go"

type AbstractBaseBroker interface {
	Init()
	GetConnection() amqp.Connection
	GetConsumers() int
	Channel()
	Consume(eventName string, callback Callback)
	Produce(events Event)
	RabbitURL() string
	CloseConnection()
}

type BaseBroker struct {
	AbstractBaseBroker
}
