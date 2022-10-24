package EventPeople

import amqp "github.com/rabbitmq/amqp091-go"

type AbstractBaseBroker interface {
    Init()
    GetConnection() amqp.Connection
    GetConsumers() int
    Consume(eventName string, callback Callback)
    Produce(events Event)
    CloseConnection()
}

type BaseBroker struct {
    AbstractBaseBroker
}
