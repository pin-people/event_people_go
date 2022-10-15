package main

import (
	"encoding/json"
	"fmt"
	"os"

	EventPeople "github.com/pinpeople/event_people_go/lib/event_people"
)

func init() {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "service")
	os.Setenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME", "event_people")
	os.Setenv("RABBIT_EVENT_PEOPLE_VHOST", "event_people")
	os.Setenv("RABBIT_URL", "amqp://admin:admin@localhost:5672")
	os.Setenv("RABBIT_FULL_URL", fmt.Sprintf("%s/%s", os.Getenv("RABBIT_URL"), os.Getenv("RABBIT_EVENT_PEOPLE_VHOST")))

	EventPeople.Config.Init()
}

type BodyStructureDaemon struct {
	Amount int    `json:"amount"`
	Name   string `json:"name"`
}

type PrivateMessageDaemon struct {
	Message string `json:"message"`
}

type SecondPrivateMessageDaemon struct {
	Bo string `json:"bo"`
	Dy string `json:"dy"`
}

type CustomEventListener struct {
	EventPeople.BaseListener
}

func (custom *CustomEventListener) pay(event EventPeople.Event) {
	var bodyDaemon = BodyStructureDaemon{}
	err := json.Unmarshal([]byte(fmt.Sprintf("%v", event.Body)), &bodyDaemon)
	EventPeople.FailOnError(err, "Error on unmarchal daemon pay")

	fmt.Println(fmt.Sprintf("Paid %v for %s ~> %s", bodyDaemon.Amount, bodyDaemon.Name, event.GetRoutingKey()))
	custom.Success()
}

func (custom *CustomEventListener) receive(event EventPeople.Event) {
	var bodyDaemon = BodyStructureDaemon{}
	err := json.Unmarshal([]byte(fmt.Sprintf("%v", event.Body)), &bodyDaemon)
	EventPeople.FailOnError(err, "Error on unmarchal daemon pay")

	if bodyDaemon.Amount < 500 {
		fmt.Println(fmt.Sprintf("[consumer] Got SKIPPED message:\n%d from %s ~> %s", bodyDaemon.Amount, bodyDaemon.Name, event.GetRoutingKey()))
		custom.Reject()
		return
	}
	fmt.Println(fmt.Sprintf("Received %d from %s ~> %s", bodyDaemon.Amount, bodyDaemon.Name, event.GetRoutingKey()))
	custom.Success()
}

func (custom *CustomEventListener) privateChannel(event EventPeople.Event) {
	var bodyDaemon = PrivateMessageDaemon{}
	err := json.Unmarshal([]byte(fmt.Sprintf("%v", event.Body)), &bodyDaemon)
	EventPeople.FailOnError(err, "Error on unmarchal daemon pay")

	fmt.Println(fmt.Sprintf("[Consumer] Got a private message: %s ~> %s", bodyDaemon.Message, event.GetRoutingKey()))
	custom.Success()
}

func (custom *CustomEventListener) secondPrivateChannel(event EventPeople.Event) {
	var bodyDaemon = SecondPrivateMessageDaemon{}
	err := json.Unmarshal([]byte(fmt.Sprintf("%v", event.Body)), &bodyDaemon)
	EventPeople.FailOnError(err, "Error on unmarchal daemon pay")

	fmt.Println(fmt.Sprintf("[Consumer] Got a private message: %s ~> %s", fmt.Sprintf("bo: %s -> dy: %s", bodyDaemon.Bo, bodyDaemon.Dy), event.GetRoutingKey()))
	custom.Success()
}

func RunDaemon() {
	custom := new(CustomEventListener)
	custom.BindEvent(custom.pay, "resource.custom.pay")
	custom.BindEvent(custom.receive, "resource.custom.receive")
	custom.BindEvent(custom.privateChannel, "resource.custom.private.service")
	custom.BindEvent(custom.secondPrivateChannel, "resource.origin.action")

	EventPeople.NewDaemon().Start()
}

func main() {
	RunDaemon()
}
