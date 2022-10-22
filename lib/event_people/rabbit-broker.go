package EventPeople

import (
	"fmt"
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

func (rabbit *RabbitBroker) Init() {
	connection, err := amqp.Dial(rabbit.RabbitURL())
	FailOnError(err, "Failed to connect to RabbitMQ")
	rabbit.connection = connection
	rabbit.topic = Topic{}
}

func (rabbit *RabbitBroker) GetConnection() amqp.Connection {
	return *rabbit.connection
}

func (rabbit *RabbitBroker) GetConsumers() int {
	return rabbit.queue.GetConsumers()
}

func (rabbit *RabbitBroker) Channel() {
	channel, err := rabbit.connection.Channel()
	FailOnError(err, "Failed to open a channel")
	rabbit.amqpChannel = channel
	rabbit.amqpChannel.Qos(1, 0, false)
	rabbit.topic.Init(rabbit.amqpChannel)
}

func (rabbit *RabbitBroker) Consume(eventName string, callback Callback) {
	if rabbit.connection == nil {
		rabbit.Init()
	}

	if rabbit.amqpChannel == nil {
		rabbit.Channel()
	}
	rabbit.queue = Queue{channel: rabbit.amqpChannel}
	rabbit.queue.Subscribe(eventName, callback)
}

func (rabbit *RabbitBroker) Produce(event Event) {
	if rabbit.connection == nil {
		rabbit.Init()
	}

	rabbit.Channel()

	rabbit.topic.Init(rabbit.amqpChannel)
	rabbit.topic.Produce(event)
}

func (rabbit *RabbitBroker) RabbitURL() string {
	return fmt.Sprintf("%s/%s", os.Getenv("RABBIT_URL"), os.Getenv("RABBIT_EVENT_PEOPLE_VHOST"))
}

func (rabbit *RabbitBroker) CloseConnection() {
	rabbit.connection.Close()
}
