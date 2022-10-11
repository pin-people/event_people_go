package RabbitContent

import (
	"encoding/json"
	"fmt"
	"strings"

	Config "github.com/pinpeople/event_people_go/lib/event_people"

	Callback "github.com/pinpeople/event_people_go/lib/event_people/callback"
	Context "github.com/pinpeople/event_people_go/lib/event_people/context"
	ContextEvent "github.com/pinpeople/event_people_go/lib/event_people/context-event"
	DeliveryInfo "github.com/pinpeople/event_people_go/lib/event_people/delivery-info"
	Event "github.com/pinpeople/event_people_go/lib/event_people/event"
	Utils "github.com/pinpeople/event_people_go/lib/event_people/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueInterface interface {
	Subscribe(routingKey string, callback Callback.Callback)
	SubscribeWithChannel(channel Context.ContextInterface, routingKey string, callback Callback.Callback)
	Init(channel Context.ContextInterface)
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

func (queue *Queue) Subscribe(routingKey string, callback Callback.Callback) {
	routingKeySplited := strings.Split(routingKey, ".")
	if len(routingKeySplited) == 3 {
		queue.createQueueAndBind(Config.APP_NAME+"-"+routingKey+".all", callback)
		queue.createQueueAndBind(Config.APP_NAME+"-"+routingKey+"."+Config.APP_NAME, callback)
	} else {
		queue.createQueueAndBind(Config.APP_NAME+"-"+routingKey, callback)
	}
}

func (queue *Queue) SubscribeWithChannel(channel *amqp.Channel, routingKey string, callback Callback.Callback) {
	queue.channel = channel

	routingKeySplited := strings.Split(routingKey, ".")
	if len(routingKeySplited) == 3 {
		queue.createQueueAndBind(routingKey+".all", callback)
		queue.createQueueAndBind(routingKey+"."+Config.APP_NAME, callback)
	} else {
		queue.createQueueAndBind(routingKey, callback)
	}
}

func (queue *Queue) GetConsumers() int {
	return queue.amqpQueue.Consumers
}

func (queue *Queue) QueueName(routingKey string) string {
	return queue.amqpQueue.Name
}

func (queue *Queue) callback(messages <-chan amqp.Delivery, callback Callback.Callback) {
	for message := range messages {
		var eventMessage Event.Event
		json.Unmarshal(message.Body, &eventMessage)

		eventMessage.Name = eventMessage.Headers.AppName
		eventMessage.SchemaVersion = eventMessage.Headers.SchemaVersion

		delivery := DeliveryInfo.DeliveryInfo{}
		delivery.Tag = string(message.ConsumerTag)

		listener := ContextEvent.BaseContextEvent{}
		listener.Initialize(message, delivery)

		callback(eventMessage, listener)
	}
}

func (queue *Queue) queueBind(routingKey string) {
	queuet, err := queue.channel.QueueDeclare(routingKey, true, false, false, false, nil)
	Utils.FailOnError(err, "Failed to declare a queue")
	queue.routingKey = routingKey
	queue.amqpQueue = &queuet
}

func (queue *Queue) exchangeBind() {
	err := queue.channel.ExchangeDeclare(Config.TOPIC, "topic", true, false, false, false, nil)
	Utils.FailOnError(err, "Failed to declare an exchange")

	queue.channel.QueueBind(queue.routingKey, queue.routingKey, Config.TOPIC, false, nil)
	Utils.FailOnError(err, "Failed to bind queue to exchange")
}

func (queue *Queue) createQueueAndBind(eventName string, callback Callback.Callback) {
	queue.queueBind(eventName)
	queue.exchangeBind()
	messages, err := queue.channel.Consume(eventName, Config.APP_NAME+"-"+eventName+"-"+Config.APP_NAME, false, false, false, false, nil)
	Utils.FailOnError(err, "Failed to consume a queue")
	go queue.callback(messages, callback)
	fmt.Printf("Event People consuming %s Queue!\n", queue.amqpQueue.Name)
}
