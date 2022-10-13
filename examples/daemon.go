package main

import (
	"encoding/json"
	"fmt"

	EventPeople "github.com/pinpeople/event_people_go/lib/event_people"
	Broker "github.com/pinpeople/event_people_go/lib/event_people/broker"
)

type BodyStructureDaemon struct {
	Amount int    `json:"amount"`
	Name   string `json:"name"`
}

type PrivateMessageDaemon struct {
	Message string `json:"message"`
}

type CustomEventListener struct {
	EventPeople.BaseListener
}

func (custom *CustomEventListener) pay(event EventPeople.Event) {
	var bodyDaemon = new(BodyStructureDaemon)
	err := json.Unmarshal([]byte(fmt.Sprintf("%v", event.Body)), &bodyDaemon)
	EventPeople.FailOnError(err, "Error on unmarchal daemon pay")

	fmt.Println(fmt.Sprintf("Paid %v for %s ~> %s", bodyDaemon.Amount, bodyDaemon.Name, event.Name))
	custom.Success()
}

func (custom *CustomEventListener) receive(event EventPeople.Event) {
	var bodyDaemon = new(BodyStructureDaemon)
	err := json.Unmarshal([]byte(fmt.Sprintf("%v", event.Body)), &bodyDaemon)
	EventPeople.FailOnError(err, "Error on unmarchal daemon pay")

	if bodyDaemon.Amount < 500 {
		fmt.Println(fmt.Sprintf("[consumer] Got SKIPPED message:\n%d from %s ~> %s", bodyDaemon.Amount, bodyDaemon.Name, event.Name))
		custom.Reject()
		return
	}
	fmt.Println("Received %d from %s ~> %s", bodyDaemon.Amount, bodyDaemon.Name, event.Name)
	custom.Success()
}

func (custom *CustomEventListener) privateChannel(event EventPeople.Event) {
	var bodyDaemon = new(PrivateMessageDaemon)
	err := json.Unmarshal([]byte(fmt.Sprintf("%v", event.Body)), &bodyDaemon)
	EventPeople.FailOnError(err, "Error on unmarchal daemon pay")

	fmt.Println(fmt.Sprintf("[Consumer] Got a private message: %s ~> %s", bodyDaemon.Message, event.Name))
	custom.Success()
}

func main() {
	EventPeople.Config.InitBroker(new(Broker.RabbitBroker))

	custom := new(CustomEventListener)
	custom.BindEvent(custom.pay, "resource.custom.pay")
	custom.BindEvent(custom.receive, "resource.custom.receive")
	custom.BindEvent(custom.privateChannel, "resource.custom.private.service")

	new(EventPeople.Daemon).Start()
}
