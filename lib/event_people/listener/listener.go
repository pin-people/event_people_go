package Listener

import (
	Broker "github.com/pinpeople/event_people_go/lib/event_people/broker"
	Callback "github.com/pinpeople/event_people_go/lib/event_people/callback"
)

type ListenerInterface interface {
	On(eventName string, callback Callback.Callback)
}

type Listener struct{}

func (listener *Listener) On(eventName string, callback Callback.Callback) {
	broker := new(Broker.RabbitBroker)
	broker.Init()
	broker.Consume(eventName, callback)
}
