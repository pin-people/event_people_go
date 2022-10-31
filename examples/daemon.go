package main

import (
	"fmt"
	"os"

	EventPeople "github.com/pinpeople/event_people_go/lib/event_people"
)

func init() {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "service_name")
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

type CustomEventListener struct {
	EventPeople.BaseListener
}

func (custom *CustomEventListener) pay(event EventPeople.Event) {
	var bodyDaemon = BodyStructureDaemon{}
	event.SetStructBody(&bodyDaemon)

	fmt.Println(fmt.Sprintf("Paid %v for %s ~> %s", bodyDaemon.Amount, bodyDaemon.Name, event.GetEventName()))
	custom.Success()
}

func (custom *CustomEventListener) receive(event EventPeople.Event) {
	var bodyDaemon = BodyStructureDaemon{}
	event.SetStructBody(&bodyDaemon)

	if bodyDaemon.Amount < 500 {
		fmt.Println(fmt.Sprintf("[consumer] Got SKIPPED message:\n%d from %s ~> %s", bodyDaemon.Amount, bodyDaemon.Name, event.GetEventName()))
		custom.Reject()
		return
	}
	fmt.Println(fmt.Sprintf("Received %d from %s ~> %s", bodyDaemon.Amount, bodyDaemon.Name, event.GetEventName()))
	custom.Success()
}

func (custom *CustomEventListener) privateChannel(event EventPeople.Event) {
	var bodyDaemon = PrivateMessageDaemon{}
	event.SetStructBody(&bodyDaemon)

	fmt.Println(fmt.Sprintf("[Consumer] Got a private message: \"%s\" ~> %s", bodyDaemon.Message, event.GetEventName()))
	custom.Success()
}

func (custom *CustomEventListener) ignoreMe(event EventPeople.Event) {
	var bodyDaemon = PrivateMessageDaemon{}
	event.SetStructBody(&bodyDaemon)

	fmt.Println("This should never be called...")
	fmt.Println(fmt.Sprintf("Spying on other sustems: \"%s\" ~> %s", bodyDaemon.Message, event.GetEventName()))
	custom.Success()
}

func RunDaemon() {
	custom := new(CustomEventListener)
	custom.BindEvent(custom.pay, "resource.*.pay")
	custom.BindEvent(custom.receive, "resource.custom.receive")
	custom.BindEvent(custom.privateChannel, "resource.custom.private.service")
	custom.BindEvent(custom.ignoreMe, "resource.custom.ignored.other_service")

	EventPeople.DaemonStart()
}

func main() {
	RunDaemon()
}
