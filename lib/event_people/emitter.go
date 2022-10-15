package EventPeople

import (
	"log"
)

type Emitter struct{}

func (emitter *Emitter) Trigger(events []*Event) {
	for index, event := range events {
		if event.Body == "" {
			log.Fatalf("MissingAttributeError: Event on position %d must have a body", index)
		}
		if event.Name == "" {
			log.Fatalf("MissingAttributeError: Event on position %d must have a name", index)
		}
		Config.Broker.Produce(*event)
	}
}

func NewEmitter() *Emitter {
	return new(Emitter)
}
