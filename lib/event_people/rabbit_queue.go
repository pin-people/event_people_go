package EventPeople

import (
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
	amqpQueue *amqp.Queue
	channel   *amqp.Channel
	QueueInterface
}

func (queue *Queue) Init(channel *amqp.Channel) {
	queue.channel = channel
}

func (queue *Queue) Subscribe(routingKey string) error {
	return queue.createQueueAndBind(routingKey)
}

func (queue *Queue) Consume(routingKey string) (<-chan amqp.Delivery, error) {
	queueName := queue.queueNameByRoutingKey(routingKey)
	queue.inspectQueue(queueName)
	queue.channel.Qos(workerPool, 0, false)
	return queue.channel.Consume(queueName, "", false, false, false, false, nil)
}

func (queue *Queue) GetConsumers() int {
	return queue.amqpQueue.Consumers
}

func (queue *Queue) QueueName(routingKey string) string {
	return queue.amqpQueue.Name
}

func (queue *Queue) createQueue(queueName string) error {
	localQueue, err := queue.channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return err
	}
	queue.amqpQueue = &localQueue
	return nil
}

func (queue *Queue) inspectQueue(queueName string) error {
	localQueue, err := queue.channel.QueueInspect(queueName)
	if err != nil {
		return err
	}
	queue.amqpQueue = &localQueue
	return nil
}

func (queue *Queue) exchangeBind(queueName string, routingKey string) error {
	err := queue.channel.ExchangeDeclarePassive(os.Getenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME"), "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = queue.channel.QueueBind(queueName, routingKey, os.Getenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME"), false, nil)

	if err != nil {
		return err
	}
	return nil
}

func (queue *Queue) createQueueAndBind(routingKey string) error {
	queueName := queue.queueNameByRoutingKey(routingKey)
	err := queue.createQueue(queueName)
	if err != nil {
		return err
	}
	err = queue.exchangeBind(queueName, routingKey)
	return err
}

func (queue *Queue) queueNameByRoutingKey(routingKey string) string {
	eventNameSplited := strings.Split(routingKey, ".")
	return os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME") + "-" + strings.Join(eventNameSplited, ".")
}
