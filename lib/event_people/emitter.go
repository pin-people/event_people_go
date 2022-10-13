package EventPeople

import (
	"fmt"
	"log"
)

type Emitter struct{}

func (emitter *Emitter) Trigger(events []*Event) {
	for index := 0; index < len(events); index++ {
		if events[index].Body == "" {
			log.Fatal(fmt.Sprintf("MissingAttributeError: Event on position %d must have a body", index))
		}
		if events[index].Name == "" {
			log.Fatal(fmt.Sprintf("MissingAttributeError: Event on position %d must have a name", index))
		}
		Config.Broker.Produce(*events[index])
	}
}
