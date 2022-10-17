package EventPeople

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueInterface interface {
	Subscribe(routingKey string, callback Callback)
	SubscribeWithChannel(channel ContextInterface, routingKey string, callback Callback)
	Init(channel ContextInterface)
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

func (queue *Queue) Subscribe(routingKey string, callback Callback) {
	routingKeySplited := strings.Split(routingKey, ".")
	if len(routingKeySplited) == 3 {
		queue.createQueueAndBind(os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME")+"-"+routingKey+".all", callback)
		queue.createQueueAndBind(os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME")+"-"+routingKey+"."+os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME"), callback)
	} else {
		queue.createQueueAndBind(os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME")+"-"+routingKey, callback)
	}
}

func (queue *Queue) SubscribeWithChannel(channel *amqp.Channel, routingKey string, callback Callback) {
	queue.channel = channel

	queue.Subscribe(routingKey, callback)
}

func (queue *Queue) GetConsumers() int {
	return queue.amqpQueue.Consumers
}

func (queue *Queue) QueueName(routingKey string) string {
	return queue.amqpQueue.Name
}

func (queue *Queue) callback(messages <-chan amqp.Delivery, callback Callback) {
	for message := range messages {
		var eventMessage Event
		json.Unmarshal(message.Body, &eventMessage)

		eventMessage.Name = eventMessage.Headers.AppName
		eventMessage.SchemaVersion = eventMessage.Headers.SchemaVersion

		callback(eventMessage, *NewBaseListener(message, DeliveryInfo{Tag: string(message.ConsumerTag)}))
	}
}

func (queue *Queue) queueBind(routingKey string) {
	queuet, err := queue.channel.QueueDeclare(routingKey, true, false, false, false, nil)
	FailOnError(err, "Failed to declare a queue")
	queue.routingKey = routingKey
	queue.amqpQueue = &queuet
}

func (queue *Queue) exchangeBind() {
	err := queue.channel.ExchangeDeclare(os.Getenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME"), "topic", true, false, false, false, nil)
	FailOnError(err, "Failed to declare an exchange")

	queue.channel.QueueBind(queue.routingKey, queue.routingKey, os.Getenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME"), false, nil)
	FailOnError(err, "Failed to bind queue to exchange")
}

func (queue *Queue) createQueueAndBind(eventName string, callback Callback) {
	queue.queueBind(eventName)
	queue.exchangeBind()
	messages, err := queue.channel.Consume(eventName, eventName, false, false, false, false, nil)
	FailOnError(err, "Failed to consume a queue")
	go queue.callback(messages, callback)
	fmt.Printf("Event People consuming %s Queue!\n", queue.amqpQueue.Name)
}
