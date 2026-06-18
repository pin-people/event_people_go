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

// amqpChannel is the slice of *amqp.Channel that Queue relies on. Declaring it as
// an interface keeps the queue-declaration path unit-testable with a fake.
type amqpChannel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	QueueInspect(name string) (amqp.Queue, error)
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error
	ExchangeDeclarePassive(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	Qos(prefetchCount, prefetchSize int, global bool) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
}

type Queue struct {
	amqpQueue *amqp.Queue
	channel   amqpChannel
	QueueInterface
}

func (queue *Queue) Init(channel amqpChannel) {
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
	// The main queue is declared WITHOUT a dead-letter-exchange argument.
	// Dead-lettering is handled at the application level (see RabbitContext:
	// Reject and retry exhaustion publish to <app>_dlq). Keeping the main queue
	// argument-free means upgrades over legacy queues never hit PRECONDITION_FAILED
	// on redeclare.
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

// declareDLQ declares the application-level dead-letter queue (idempotent).
// It is a plain durable queue, not bound to any dead-letter-exchange: the library
// publishes failed messages to it directly (see RabbitContext). No DLX fanout
// exchange or binding is created, so there is no broker-side topology to drift
// between library versions.
func (queue *Queue) declareDLQ() error {
	appName := os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME")
	dlqName := appName + "_dlq"

	_, err := queue.channel.QueueDeclare(dlqName, true, false, false, false, nil)
	return err
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

	// Declare the application-level DLQ first (idempotent)
	err := queue.declareDLQ()
	if err != nil {
		return err
	}

	// Declare retry queue (idempotent)
	err = queue.declareRetryQueue(queueName)
	if err != nil {
		return err
	}

	// Declare the main queue (argument-free; dead-lettering is app-level)
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
