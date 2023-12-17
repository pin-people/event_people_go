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

func (rabbit *RabbitBroker) Consume(eventName string, callback Callback) {
	if rabbit.connection == nil {
		rabbit.Init()
	}

	if rabbit.amqpChannel == nil {
		rabbit.Channel()
	}
	channel, err := rabbit.connection.Channel()
	queue := Queue{channel: channel}
	deliveries, err := queue.Consume(eventName)

	if err != nil {
		log.Println(err)
	}
	for delivery := range deliveries {
		var eventMessage Event
		json.Unmarshal(delivery.Body, &eventMessage)

		eventMessage.Name = eventMessage.Headers.AppName
		eventMessage.SchemaVersion = eventMessage.Headers.SchemaVersion
		deliveryStruct := DeliveryStruct{DeliveryInterface: delivery, Body: delivery.Body, DeliveryTag: delivery.DeliveryTag}
		rabbitContext := NewContext(delivery)
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
