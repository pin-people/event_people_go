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
	err := queue.inspectQueue(queueName)
	if err != nil {
		return nil, err
	}
	err = queue.channel.Qos(workerPool, 0, false)
	if err != nil {
		return nil, err
	}
	return queue.channel.Consume(queueName, "", false, false, false, false, nil)
}

func (queue *Queue) GetConsumers() int {
	return queue.amqpQueue.Consumers
}

func (queue *Queue) QueueName(routingKey string) string {
	return queue.amqpQueue.Name
}

func (queue *Queue) createQueue(queueName string) error {
	appName := os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME")
	dlxName := appName + "_dlx"

	args := amqp.Table{
		"x-dead-letter-exchange": dlxName,
	}
	localQueue, err := queue.channel.QueueDeclare(queueName, true, false, false, false, args)
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

// declareDLXTopology declares the DLX exchange and DLQ (idempotent).
func (queue *Queue) declareDLXTopology() error {
	appName := os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME")
	dlxName := appName + "_dlx"
	dlqName := appName + "_dlq"

	// Declare DLX as a fanout exchange
	err := queue.channel.ExchangeDeclare(dlxName, "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}

	// Declare DLQ
	_, err = queue.channel.QueueDeclare(dlqName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	// Bind DLQ to DLX with empty routing key
	err = queue.channel.QueueBind(dlqName, "", dlxName, false, nil)
	if err != nil {
		return err
	}

	return nil
}

// declareRetryQueue declares the retry queue for a given main queue name (idempotent).
// The retry queue uses a per-message TTL (set via Expiration in Publishing) and a
// dead-letter-routing-key back to the original queue so messages re-enter after delay.
func (queue *Queue) declareRetryQueue(queueName string) error {
	retryQueueName := queueName + "_retry"
	args := amqp.Table{
		"x-dead-letter-exchange":    "",        // default exchange
		"x-dead-letter-routing-key": queueName, // route back to original queue
		// NOTE: x-message-ttl is intentionally NOT set here;
		// per-message TTL is controlled via the Expiration field of each Publishing.
	}
	_, err := queue.channel.QueueDeclare(retryQueueName, true, false, false, false, args)
	return err
}

func (queue *Queue) createQueueAndBind(routingKey string) error {
	queueName := queue.queueNameByRoutingKey(routingKey)

	// Declare DLX topology first (idempotent)
	err := queue.declareDLXTopology()
	if err != nil {
		return err
	}

	// Declare retry queue (idempotent)
	err = queue.declareRetryQueue(queueName)
	if err != nil {
		return err
	}

	// Declare main queue with dead-letter-exchange
	err = queue.createQueue(queueName)
	if err != nil {
		return err
	}

	err = queue.exchangeBind(queueName, routingKey)
	return err
}

// queueNameByRoutingKey returns '{appName}-{resource}.{origin}.{action}.all'.
// It always normalizes the destination segment to 'all', regardless of input
// routing key — per spec routing_convention note and DEV-GO-002 fix.
func (queue *Queue) queueNameByRoutingKey(routingKey string) string {
	eventNameSplited := strings.Split(routingKey, ".")
	if len(eventNameSplited) <= 3 {
		eventNameSplited = append(eventNameSplited, "all")
	} else {
		// Normalize the destination segment (index 3) to "all".
		// The exchange binding still uses the original routing key for correct routing.
		eventNameSplited[3] = "all"
	}
	return os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME") + "-" + strings.Join(eventNameSplited, ".")
}
