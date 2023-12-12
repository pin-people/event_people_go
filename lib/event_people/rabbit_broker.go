package EventPeople

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueInfo struct {
	Name     string                 `json:"name"`
	Messages int                    `json:"messages"`
	Args     map[string]interface{} `json:"arguments"`
}

type RabbitBroker struct {
	queue       Queue
	topic       Topic
	connection  *amqp.Connection
	amqpChannel *amqp.Channel
	queuesInfo  []QueueInfo
	*BaseBroker
}

func (rabbit *RabbitBroker) Init() {
	connection, err := amqp.Dial(rabbit.RabbitURL())
	FailOnError(err, "Failed to connect to RabbitMQ")
	rabbit.connection = connection
	rabbit.topic = Topic{}
	rabbit.queuesInfo = rabbit.getQueuesInformation()
}

func (rabbit *RabbitBroker) GetConnection() amqp.Connection {
	return *rabbit.connection
}

func (rabbit *RabbitBroker) GetConsumers() int {
	return rabbit.queue.GetConsumers()
}

func (rabbit *RabbitBroker) Channel() {
	workerPool, _ := strconv.Atoi(os.Getenv("WORKERS"))
	if workerPool == 0 {
		workerPool = runtime.NumCPU() * 2
	}
	channel, err := rabbit.connection.Channel()
	FailOnError(err, "Failed to open a channel")
	rabbit.amqpChannel = channel
	rabbit.amqpChannel.Qos(workerPool, 0, false)
	rabbit.topic.Init(rabbit.amqpChannel)
}

func (rabbit *RabbitBroker) Subscribe(eventName string) {
	if rabbit.connection == nil {
		rabbit.Init()
	}

	if rabbit.amqpChannel == nil {
		rabbit.Channel()
	}
	rabbit.queue = Queue{
		channel:   rabbit.amqpChannel,
		queueInfo: rabbit.queuesInfo,
	}
	if eventName == dlxEventName {
		rabbit.queue.CreateDLX()
		return
	}
	rabbit.queue.Subscribe(eventName)

}

func (rabbit *RabbitBroker) Consume(eventName string) *DeliveryStruct {
	if rabbit.connection == nil {
		rabbit.Init()
	}

	if rabbit.amqpChannel == nil {
		rabbit.Channel()
	}
	rabbit.queue = Queue{
		channel:   rabbit.amqpChannel,
		queueInfo: rabbit.queuesInfo,
	}
	delivery := rabbit.queue.Consume(eventName)
	if delivery == nil {
		return nil
	}
	return &DeliveryStruct{DeliveryInterface: delivery, Body: delivery.Body, DeliveryTag: delivery.DeliveryTag}
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

func (rabbit *RabbitBroker) getQueuesInformation() []QueueInfo {
	rabbitUrl := strings.ReplaceAll(os.Getenv("RABBIT_URL"), "amqp://", "")
	splittedRabbitUrl := strings.Split(rabbitUrl, "@")
	usernameAndPassword := strings.Split(splittedRabbitUrl[0], ":")
	username := usernameAndPassword[0]
	password := usernameAndPassword[1]
	splittedHost := strings.Split(splittedRabbitUrl[1], ":")
	host := splittedHost[0]
	rabbitMQURL := fmt.Sprintf("http://%s:15672/api/queues/%s", host, os.Getenv("RABBIT_EVENT_PEOPLE_VHOST"))

	req, err := http.NewRequest("GET", rabbitMQURL, nil)
	if err != nil {
		log.Fatalf("Erro ao criar a solicitação HTTP: %s", err)
	}

	req.SetBasicAuth(username, password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("HTTP Request error: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Fail getting queue informations: %s", resp.Status)
	}

	var queues []QueueInfo
	err = json.NewDecoder(resp.Body).Decode(&queues)
	if err != nil {
		log.Fatalf("Error on json decode JSON: %s", err)
	}
	return queues
}
