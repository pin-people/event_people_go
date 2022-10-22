package EventPeople

import (
	"log"
)

func TriggerEmitter(events []*Event) {
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
