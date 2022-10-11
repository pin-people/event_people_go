package Models

import (
	"fmt"
	"log"

	Broker "github.com/pinpeople/event_people_go/lib/event_people/broker"
	Event "github.com/pinpeople/event_people_go/lib/event_people/event"
)

type Emitter struct{}

func (emitter *Emitter) Trigger(events []*Event.Event, broker Broker.RabbitBroker) {
	for index := 0; index < len(events); index++ {
		if events[index].Body == "" {
			log.Fatal(fmt.Sprintf("MissingAttributeError: Event on position %d must have a body", index))
		}
		if events[index].Name == "" {
			log.Fatal(fmt.Sprintf("MissingAttributeError: Event on position %d must have a name", index))
		}
		broker.Produce(*events[index])
	}
}
