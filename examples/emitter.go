package main

import (
	"bytes"
	"encoding/json"

	EventPeople "github.com/pinpeople/event_people_go/lib/event_people"
	Broker "github.com/pinpeople/event_people_go/lib/event_people/broker"
)

type BodyStructureEmmiter struct {
	Amount int    `json:"amount"`
	Name   string `json:"name"`
}

type PrivateMessageEmitter struct {
	Message string `json:"message"`
}

func RunEmitter() {
	EventPeople.Config.InitBroker(new(Broker.RabbitBroker))

	var events []*EventPeople.Event

	firstBody, err := json.Marshal(BodyStructureEmmiter{Amount: 1500, Name: "John"})

	EventPeople.FailOnError(err, "error on create body")

	firstItem := new(EventPeople.Event)
	firstItem.Initialize("service-resource.custom.pay", bytes.NewBuffer(firstBody).String())

	events = append(events, firstItem)

	secondBody, err := json.Marshal(BodyStructureEmmiter{Amount: 35, Name: "Peter"})
	EventPeople.FailOnError(err, "error on create body")

	secondItem := new(EventPeople.Event)
	secondItem.Initialize("service-resource.custom.receive", bytes.NewBuffer(secondBody).String())

	events = append(events, secondItem)

	thirdBody, err := json.Marshal(BodyStructureEmmiter{Amount: 350, Name: "George"})
	EventPeople.FailOnError(err, "error on create body")

	thirdItem := new(EventPeople.Event)
	thirdItem.Initialize("service-resource.custom.receive", bytes.NewBuffer(thirdBody).String())

	events = append(events, thirdItem)

	fourthBody, err := json.Marshal(BodyStructureEmmiter{Amount: 550, Name: "James"})
	EventPeople.FailOnError(err, "error on create body")

	fourthItem := new(EventPeople.Event)
	fourthItem.Initialize("service-resource.custom.receive", bytes.NewBuffer(fourthBody).String())

	events = append(events, fourthItem)

	singleBody, err := json.Marshal(PrivateMessageEmitter{Message: "Secret"})
	EventPeople.FailOnError(err, "error on create body")

	singleItem := new(EventPeople.Event)
	singleItem.Initialize("service-resource.custom.private.service", bytes.NewBuffer(singleBody).String())

	events = append(events, singleItem)

	actionBody, err := json.Marshal(BodyStructureEmmiter{Amount: 30, Name: "Willian"})
	EventPeople.FailOnError(err, "error on create body")

	actionItem := new(EventPeople.Event)
	actionItem.Initialize("service-resource.origin.action", bytes.NewBuffer(actionBody).String())

	events = append(events, actionItem)

	new(EventPeople.Emitter).Trigger(events)
}

func main() {
	RunEmitter()
}
