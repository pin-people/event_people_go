package Broker

import (
	Callback "github.com/pinpeople/event_people_go/lib/event_people/callback"
	Event "github.com/pinpeople/event_people_go/lib/event_people/event"
)

type AbstractBase interface {
	GetConnection()
	GetConsumers() int
	Channel()
	Consume(eventName string, callback Callback.Callback)
	Produce(events Event.Event)
	RabbitURL() string
	CloseConnection()
}

type BaseBroker struct {
	AbstractBase
}
