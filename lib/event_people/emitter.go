package EventPeople

import (
	"log"
)

func TriggerEmitter(events []*Event) {
	for index, event := range events {
		if event.Body == "" {
			log.Println("MissingAttributeError: Event on position %d must have a body", index)
		}
		if event.Name == "" {
			log.Println("MissingAttributeError: Event on position %d must have a name", index)
		}
	}
	for _, event := range events {
		Config.Broker.Produce(*event)
	}
}
