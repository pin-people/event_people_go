package EventPeople

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitBroker struct {
	queue       Queue
	topic       Topic
	connection  *amqp.Connection
	amqpChannel *amqp.Channel
	*BaseBroker
}

func (rabbit *RabbitBroker) Init() error {
	connection, err := amqp.Dial(rabbit.RabbitURL())
	if err != nil {
		return err
	}
	rabbit.connection = connection
	rabbit.topic = Topic{}
	return nil
}

func (rabbit *RabbitBroker) GetConnection() amqp.Connection {
	return *rabbit.connection
}

func (rabbit *RabbitBroker) GetConsumers() int {
	return rabbit.queue.GetConsumers()
}

func (rabbit *RabbitBroker) Channel() error {
	channel, err := rabbit.connection.Channel()
	if err != nil {
		return err
	}
	rabbit.amqpChannel = channel
	rabbit.amqpChannel.Qos(1, 0, false)
	rabbit.topic.Init(rabbit.amqpChannel)
	return nil
}

func (rabbit *RabbitBroker) Subscribe(eventName string) error {
	if rabbit.connection == nil {
		rabbit.Init()
	}

	if rabbit.amqpChannel == nil {
		rabbit.Channel()
	}
	channel, err := rabbit.connection.Channel()
	defer channel.Close()
	queue := Queue{channel: channel}
	if err != nil {
		log.Println(err)
	}
	err = queue.Subscribe(eventName)
	return err
}

func (rabbit *RabbitBroker) Consume(eventName string, callback Callback, retryConfig ...RetryConfig) {
	if rabbit.connection == nil {
		rabbit.Init()
	}

	if rabbit.amqpChannel == nil {
		rabbit.Channel()
	}
	channel, err := rabbit.connection.Channel()
	queue := Queue{channel: channel}

	go func() {
		for {
			_, ok := <-channel.NotifyClose(make(chan *amqp.Error))
			if !ok {
				newChannel, err := rabbit.connection.Channel()
				if err == nil {
					channel = newChannel
				}
			}
		}
	}()

	deliveries, err := queue.Consume(eventName)

	if err != nil {
		log.Println(err)
	}

	// Resolve retry config — use provided override or fall back to global Config defaults.
	var rc RetryConfig
	if len(retryConfig) > 0 {
		rc = retryConfig[0]
	} else {
		rc = Config.GetRetryConfig()
	}

	queueName := queue.queueNameByRoutingKey(eventName)

	for delivery := range deliveries {
		var eventMessage Event
		json.Unmarshal(delivery.Body, &eventMessage)

		eventMessage.Name = delivery.RoutingKey
		eventMessage.SchemaVersion = eventMessage.Headers.SchemaVersion

		// Read retry count from AMQP header
		retryCount := 0
		if v, ok := delivery.Headers["x-event-people-retries"]; ok {
			retryCount = int(v.(int32))
		}
		eventMessage.RetryCount = retryCount

		deliveryStruct := DeliveryStruct{
			DeliveryInterface: delivery,
			Body:              delivery.Body,
			DeliveryTag:       delivery.DeliveryTag,
			RoutingKey:        delivery.RoutingKey,
			Headers:           headersToMap(delivery.Headers),
			ContentType:       delivery.ContentType,
			MaxRetries:        rc.MaxAttempts,
			DelayStrategy:     rc.DelayStrategy,
			QueueName:         queueName,
			RetryCount:        retryCount,
		}

		rabbitContext := NewContextWithRetry(
			delivery,
			channel,
			queueName,
			rc.MaxAttempts,
			rc.DelayStrategy,
			retryCount,
		)
		rabbitContext.DeliveryStruct = deliveryStruct

		callback(eventMessage, rabbitContext)
	}
}

func (rabbit *RabbitBroker) Produce(event Event) error {
	if rabbit.connection == nil {
		rabbit.Init()
	}

	rabbit.Channel()

	rabbit.topic.Init(rabbit.amqpChannel)
	return rabbit.topic.Produce(event)
}

func (rabbit *RabbitBroker) RabbitURL() string {
	return fmt.Sprintf("%s/%s", os.Getenv("RABBIT_URL"), os.Getenv("RABBIT_EVENT_PEOPLE_VHOST"))
}

func (rabbit *RabbitBroker) CloseConnection() {
	rabbit.connection.Close()
}

// headersToMap converts an amqp.Table to map[string]interface{}.
func headersToMap(table amqp.Table) map[string]interface{} {
	m := make(map[string]interface{}, len(table))
	for k, v := range table {
		m[k] = v
	}
	return m
}
