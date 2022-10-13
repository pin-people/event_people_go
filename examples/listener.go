package main

import (
	"encoding/json"
	"fmt"

	EventPeople "github.com/pinpeople/event_people_go/lib/event_people"
	Broker "github.com/pinpeople/event_people_go/lib/event_people/broker"
)

func defaultCallback(event EventPeople.Event, context EventPeople.BaseListener) {
	msg, err := json.Marshal(event.Body)
	EventPeople.FailOnError(err, "Error on received event")
	fmt.Println(
		fmt.Sprintf("EventName: %s \nBody: %s", event.GetRoutingKey(), string(msg)),
	)
	context.Success()
}

func RunListener() {
	EventPeople.ConfigEnvs()
	EventPeople.Config.InitBroker(new(Broker.RabbitBroker))
	new(EventPeople.Listener).On("resource.custom.*", defaultCallback)
	new(EventPeople.Listener).On("resource.origin.*", defaultCallback)
}

// func main() {
// 	RunListener()
// }
