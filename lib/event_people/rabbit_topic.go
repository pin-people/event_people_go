package EventPeople

import (
	"context"
	"fmt"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Topic struct {
	channel *amqp.Channel
}

func (topic *Topic) Init(channel *amqp.Channel) {
	topic.channel = channel
}

func (topic *Topic) Produce(event Event) error {
	payload, err := event.Payload()
	if err != nil {
		return err
	}
	message := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "text/plain",
		Body:         []byte(payload),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	fmt.Printf("Producing message to %s!\n", event.GetEventName())
	err = topic.channel.PublishWithContext(ctx, os.Getenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME"), event.GetEventName(), false, false, message)
	if err != nil {
		return err
	}
	fmt.Printf("Message sent to %s!\n", event.GetEventName())
	defer cancel()
	return nil
}

func (topic *Topic) GetChannel() *amqp.Channel {
	return topic.channel
}
