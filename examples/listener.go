package main

import (
	"encoding/json"
	"fmt"

	Config "github.com/pinpeople/event_people_go/lib/event_people"
	ContextEvent "github.com/pinpeople/event_people_go/lib/event_people/context-event"
	Event "github.com/pinpeople/event_people_go/lib/event_people/event"
	Listener "github.com/pinpeople/event_people_go/lib/event_people/listener"
	Utils "github.com/pinpeople/event_people_go/lib/event_people/utils"
)

func defaultCallback(event Event.Event, context ContextEvent.BaseContextEvent) {
	msg, err := json.Marshal(event.Body)
	Utils.FailOnError(err, "Error on received event")
	fmt.Println(
		fmt.Sprintf("EventName: %s \nBody: %s", event.GetRoutingKey(), string(msg)),
	)
	context.Success()
}

func RunListener() {
	var forever chan struct{}
	Config.ConfigEnvs()
	new(Listener.Listener).On("resource.custom.*", defaultCallback)
	new(Listener.Listener).On("resource.origin.*", defaultCallback)
	<-forever
}

// func main() {
// 	RunListener()
// }
