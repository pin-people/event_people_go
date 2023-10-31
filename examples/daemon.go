package main

import (
	"fmt"
	"os"
	"time"

	EventPeople "github.com/pin-people/event_people_go/lib/event_people"
)

func init() {
	os.Setenv("WORKERS", "4")
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

func pay(event EventPeople.Event, custom EventPeople.ContextInterface) {
	var bodyDaemon = BodyStructureDaemon{}
	event.SetStructBody(&bodyDaemon)
	fmt.Println(fmt.Sprintf("%v : Paid %v for %s ~> %s", time.Now().Format("2006-01-02 15:04:05"), bodyDaemon.Amount, bodyDaemon.Name, event.GetEventName()))
	custom.Success()
}

func receive(event EventPeople.Event, custom EventPeople.ContextInterface) {
	var bodyDaemon = BodyStructureDaemon{}
	event.SetStructBody(&bodyDaemon)

	if bodyDaemon.Amount < 500 {
		fmt.Println(fmt.Sprintf("%v : [consumer] Got SKIPPED message:\n%d from %s ~> %s", time.Now().Format("2006-01-02 15:04:05"), bodyDaemon.Amount, bodyDaemon.Name, event.GetEventName()))
		custom.Reject()
		return
	}
	fmt.Println(fmt.Sprintf("%v : Received %d from %s ~> %s", time.Now().Format("2006-01-02 15:04:05"), bodyDaemon.Amount, bodyDaemon.Name, event.GetEventName()))
	custom.Success()
}

func privateChannel(event EventPeople.Event, custom EventPeople.ContextInterface) {
	var bodyDaemon = PrivateMessageDaemon{}
	event.SetStructBody(&bodyDaemon)

	fmt.Println(fmt.Sprintf("%v : [Consumer] Got a private message: \"%s\" ~> %s", time.Now().Format("2006-01-02 15:04:05"), bodyDaemon.Message, event.GetEventName()))
	custom.Success()
}

func ignoreMe(event EventPeople.Event, custom EventPeople.ContextInterface) {
	var bodyDaemon = PrivateMessageDaemon{}
	event.SetStructBody(&bodyDaemon)

	fmt.Println("This should never be called...")
	fmt.Println(fmt.Sprintf("Spying on other sustems: \"%s\" ~> %s", bodyDaemon.Message, event.GetEventName()))
	custom.Success()
}

func RunDaemon() {
	EventPeople.BindEvent(pay, "resource.*.pay")
	EventPeople.BindEvent(receive, "resource.custom.receive")
	EventPeople.BindEvent(privateChannel, "resource.custom.private.service")
	EventPeople.BindEvent(ignoreMe, "resource.custom.ignored.other_service")

	EventPeople.DaemonStart()
}

func main() {
	RunDaemon()
}
