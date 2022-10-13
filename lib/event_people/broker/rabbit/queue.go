package RabbitContent

import (
	"encoding/json"
	"fmt"
	"strings"

	EventPeople "github.com/pinpeople/event_people_go/lib/event_people"
	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueInterface interface {
	Subscribe(routingKey string, callback EventPeople.Callback)
	SubscribeWithChannel(channel EventPeople.ContextInterface, routingKey string, callback EventPeople.Callback)
	Init(channel EventPeople.ContextInterface)
	QueueOptions()
	QueueName(routingKey string)
	queueBind()
	exchangeBind()
	callback()
}

type Queue struct {
	amqpQueue  *amqp.Queue
	channel    *amqp.Channel
	routingKey string
	QueueInterface
}

func (queue *Queue) Init(channel *amqp.Channel) {
	queue.channel = channel
}

func (queue *Queue) Subscribe(routingKey string, callback EventPeople.Callback) {
	routingKeySplited := strings.Split(routingKey, ".")
	if len(routingKeySplited) == 3 {
		queue.createQueueAndBind(EventPeople.Config.APP_NAME+"-"+routingKey+".all", callback)
		queue.createQueueAndBind(EventPeople.Config.APP_NAME+"-"+routingKey+"."+EventPeople.Config.APP_NAME, callback)
	} else {
		queue.createQueueAndBind(EventPeople.Config.APP_NAME+"-"+routingKey, callback)
	}
}

func (queue *Queue) SubscribeWithChannel(channel *amqp.Channel, routingKey string, callback EventPeople.Callback) {
	queue.channel = channel

	queue.Subscribe(routingKey, callback)
}

func (queue *Queue) GetConsumers() int {
	return queue.amqpQueue.Consumers
}

func (queue *Queue) QueueName(routingKey string) string {
	return queue.amqpQueue.Name
}

func (queue *Queue) callback(messages <-chan amqp.Delivery, callback EventPeople.Callback) {
	for message := range messages {
		var eventMessage EventPeople.Event
		json.Unmarshal(message.Body, &eventMessage)

		eventMessage.Name = eventMessage.Headers.AppName
		eventMessage.SchemaVersion = eventMessage.Headers.SchemaVersion

		delivery := EventPeople.DeliveryInfo{}
		delivery.Tag = string(message.ConsumerTag)

		listener := new(EventPeople.BaseListener)
		listener.Initialize(message, delivery)

		callback(eventMessage, *listener)
	}
}

func (queue *Queue) queueBind(routingKey string) {
	queuet, err := queue.channel.QueueDeclare(routingKey, true, false, false, false, nil)
	EventPeople.FailOnError(err, "Failed to declare a queue")
	queue.routingKey = routingKey
	queue.amqpQueue = &queuet
}

func (queue *Queue) exchangeBind() {
	err := queue.channel.ExchangeDeclare(EventPeople.Config.TOPIC, "topic", true, false, false, false, nil)
	EventPeople.FailOnError(err, "Failed to declare an exchange")

	queue.channel.QueueBind(queue.routingKey, queue.routingKey, EventPeople.Config.TOPIC, false, nil)
	EventPeople.FailOnError(err, "Failed to bind queue to exchange")
}

func (queue *Queue) createQueueAndBind(eventName string, callback EventPeople.Callback) {
	queue.queueBind(eventName)
	queue.exchangeBind()
	messages, err := queue.channel.Consume(eventName, EventPeople.Config.APP_NAME+"-"+eventName+"-"+EventPeople.Config.APP_NAME, false, false, false, false, nil)
	EventPeople.FailOnError(err, "Failed to consume a queue")
	go queue.callback(messages, callback)
	fmt.Printf("Event People consuming %s Queue!\n", queue.amqpQueue.Name)
}
