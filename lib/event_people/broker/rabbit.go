package Broker

import (
	Config "github.com/pinpeople/event_people_go/lib/event_people"
	RabbitContent "github.com/pinpeople/event_people_go/lib/event_people/broker/rabbit"
	Callback "github.com/pinpeople/event_people_go/lib/event_people/callback"
	Event "github.com/pinpeople/event_people_go/lib/event_people/event"
	Utils "github.com/pinpeople/event_people_go/lib/event_people/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	connection  *amqp.Connection
	amqpChannel *amqp.Channel
}

type RabbitBroker struct {
	queue RabbitContent.Queue
	topic RabbitContent.Topic
	RabbitMQ
	AbstractBase
}

func (rabbit *RabbitBroker) Init() {
	connection, err := amqp.Dial(Config.FULL_URL)
	Utils.FailOnError(err, "Failed to connect to RabbitMQ")
	rabbit.connection = connection
	rabbit.topic = RabbitContent.Topic{}
}

func (rabbit *RabbitBroker) GetConnection() amqp.Connection {
	return *rabbit.connection
}

func (rabbit *RabbitBroker) GetConsumers() int {
	return rabbit.queue.GetConsumers()
}

func (rabbit *RabbitBroker) Channel() {
	channel, err := rabbit.connection.Channel()
	Utils.FailOnError(err, "Failed to open a channel")
	rabbit.amqpChannel = channel
	rabbit.topic.Init(rabbit.amqpChannel)
}

func (rabbit *RabbitBroker) Consume(eventName string, callback Callback.Callback) {
	if rabbit.connection == nil {
		rabbit.Init()
	}

	if rabbit.amqpChannel == nil {
		rabbit.Channel()
	}
	rabbit.queue = RabbitContent.Queue{}
	rabbit.queue.SubscribeWithChannel(rabbit.amqpChannel, eventName, callback)
}

func (rabbit *RabbitBroker) Produce(event Event.Event) {
	if rabbit.connection == nil {
		rabbit.Init()
	}

	rabbit.Channel()

	rabbit.topic.Init(rabbit.amqpChannel)
	rabbit.topic.Produce(event)
}

func (rabbit *RabbitBroker) RabbitURL() string {
	return Config.FULL_URL
}

func (rabbit *RabbitBroker) CloseConnection() {
	rabbit.connection.Close()
}
