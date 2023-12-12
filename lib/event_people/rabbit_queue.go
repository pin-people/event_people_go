package EventPeople

import (
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
	amqpQueue *amqp.Queue
	channel   *amqp.Channel
	queueInfo []QueueInfo
	QueueInterface
}

func (queue *Queue) Init(channel *amqp.Channel) {
	queue.channel = channel
}

func (queue *Queue) Subscribe(routingKey string) {
	queue.createQueueAndBind(routingKey)
}

func (queue *Queue) Consume(routingKey string) *amqp.Delivery {
	queueName := queue.queueNameByRoutingKey(routingKey)
	queue.inspectQueue(queueName)
	delivery, ok, err := queue.channel.Get(queueName, false)
	FailOnError(err, "Failed to consume a queue")
	if ok {
		return &delivery
	}
	return nil
}

func (queue *Queue) GetConsumers() int {
	return queue.amqpQueue.Consumers
}

func (queue *Queue) QueueName(routingKey string) string {
	return queue.amqpQueue.Name
}

func (queue *Queue) CreateDLX() {
	exchangeDlxName := queue.getExchangeDlxName()
	dlxQueueName := queue.getQueueDlxName()
	_, err := queue.channel.QueueDeclare(dlxQueueName, true, false, false, false, nil)
	FailOnError(err, "Failed to declare a queue")

	err = queue.channel.ExchangeDeclare(exchangeDlxName, "topic", true, false, false, false, nil)
	FailOnError(err, "Failed to declare an exchange")

	dlxRoutingKey := queue.getRoutingKeyDlxName()
	queue.channel.QueueBind(dlxQueueName, dlxRoutingKey, exchangeDlxName, false, nil)
}

func (queue *Queue) createQueue(queueName string) {
	args := amqp.Table{}
	if Config.UseDLX {
		exchangeDlxName := queue.getExchangeDlxName()
		dlxRoutingKey := queue.getRoutingKeyDlxName()
		args = amqp.Table{
			"x-dead-letter-exchange":    exchangeDlxName,
			"x-dead-letter-routing-key": dlxRoutingKey,
		}
	}

	for _, item := range queue.queueInfo {
		if item.Name == queueName {
			if len(item.Args) != len(args) {
				break
			}
			for key, value := range args {
				if item.Args[key] != value {
					break
				}
			}
			return
		}
	}

	inspectedQueue, err := queue.channel.QueueInspect(queueName)
	if inspectedQueue.Messages > 0 {
		return
	}

	queue.channel.QueueDelete(queueName, false, false, false)
	localQueue, err := queue.channel.QueueDeclare(queueName, true, false, false, false, args)
	FailOnError(err, "Failed to declare a queue")
	queue.amqpQueue = &localQueue
}

func (queue *Queue) inspectQueue(queueName string) {
	localQueue, err := queue.channel.QueueInspect(queueName)
	FailOnError(err, "Failed to declare a queue")
	queue.amqpQueue = &localQueue
}

func (queue *Queue) exchangeBind(queueName string, routingKey string) {
	err := queue.channel.ExchangeDeclare(os.Getenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME"), "topic", true, false, false, false, nil)
	FailOnError(err, "Failed to declare an exchange")

	queue.channel.QueueBind(queueName, routingKey, os.Getenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME"), false, nil)
	FailOnError(err, "Failed to bind queue to exchange")
}

func (queue *Queue) createQueueAndBind(routingKey string) {
	queueName := queue.queueNameByRoutingKey(routingKey)
	queue.createQueue(queueName)
	queue.exchangeBind(queueName, routingKey)
}

func (queue *Queue) queueNameByRoutingKey(routingKey string) string {
	eventNameSplited := strings.Split(routingKey, ".")
	return os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME") + "-" + strings.Join(eventNameSplited, ".")
}

func (queue *Queue) getExchangeDlxName() string {
	return fmt.Sprintf("%s_DLX", os.Getenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME"))
}

func (queue *Queue) getQueueDlxName() string {
	return fmt.Sprintf("%s.%s.dlx", os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME"), os.Getenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME"))
}

func (queue *Queue) getRoutingKeyDlxName() string {
	return fmt.Sprintf("%s.%s.dlx", os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME"), os.Getenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME"))
}
