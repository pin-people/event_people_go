package RabbitContent

import (
	"context"
	"fmt"
	"time"

	Config "github.com/pinpeople/event_people_go/lib/event_people"
	Event "github.com/pinpeople/event_people_go/lib/event_people/event"
	Utils "github.com/pinpeople/event_people_go/lib/event_people/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Topic struct {
	channel *amqp.Channel
}

func (topic *Topic) Init(channel *amqp.Channel) {
	topic.channel = channel
}

func (topic *Topic) Produce(event Event.Event) {
	message := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "text/plain",
		Body:         []byte(event.Payload()),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	fmt.Printf("Producing message to %s!\n", event.Name)
	err := topic.channel.PublishWithContext(ctx, Config.TOPIC, event.Name, false, false, message)
	fmt.Printf("Message to %s sended!\n", event.Name)

	Utils.FailOnError(err, "Error on publish message")
}

func (topic *Topic) GetChannel() *amqp.Channel {
	return topic.channel
}
