package main

import (
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

type BodyStructureEmmiter struct {
	Amount int    `json:"amount"`
	Name   string `json:"name"`
}

type PrivateMessageEmitter struct {
	Message string `json:"message"`
}

type SecondPrivateMessageEmitter struct {
	Bo string `json:"bo"`
	Dy string `json:"dy"`
}

func RunEmitter() {
	var events []*EventPeople.Event

	events = append(events, EventPeople.NewEvent("resource.custom.pay", BodyStructureEmmiter{Amount: 1500, Name: "John"}))
	events = append(events, EventPeople.NewEvent("resource.custom.receive", BodyStructureEmmiter{Amount: 35, Name: "Peter"}))
	events = append(events, EventPeople.NewEvent("resource.custom.receive", BodyStructureEmmiter{Amount: 350, Name: "George"}))
	events = append(events, EventPeople.NewEvent("resource.custom.receive", BodyStructureEmmiter{Amount: 550, Name: "James"}))
	events = append(events, EventPeople.NewEvent("resource.custom.private.service", PrivateMessageEmitter{Message: "Secret"}))
	EventPeople.TriggerEmitter(events)

	singleEvent := EventPeople.NewEvent("resource.origin.action", SecondPrivateMessageEmitter{Bo: "teste bo", Dy: "teste dy"})
	EventPeople.TriggerEmitter([]*EventPeople.Event{singleEvent})
	EventPeople.Config.CloseConnection()
}

func main() {
	RunEmitter()
}
