package main

import (
	"bytes"
	"encoding/json"

	Config "github.com/pinpeople/event_people_go/lib/event_people"
	Broker "github.com/pinpeople/event_people_go/lib/event_people/broker"
	Event "github.com/pinpeople/event_people_go/lib/event_people/event"
	Models "github.com/pinpeople/event_people_go/lib/event_people/models"
	Utils "github.com/pinpeople/event_people_go/lib/event_people/utils"
)

type BodyStructureOne struct {
	Amount int    `json:"amount"`
	Name   string `json:"name"`
}

func RunEmitter() {
	var forever chan struct{}
	Config.ConfigEnvs()
	rabbit := new(Broker.RabbitBroker)
	rabbit.Init()

	var events []*Event.Event

	firstBody, err := json.Marshal(BodyStructureOne{Amount: 1500, Name: "John"})

	Utils.FailOnError(err, "error on create body")

	firstItem := new(Event.Event)
	firstItem.Initialize("resource.custom.pay", bytes.NewBuffer(firstBody).String())

	events = append(events, firstItem)

	secondBody, err := json.Marshal(BodyStructureOne{Amount: 35, Name: "Peter"})
	Utils.FailOnError(err, "error on create body")

	secondItem := new(Event.Event)
	secondItem.Initialize("resource.custom.receive", bytes.NewBuffer(secondBody).String())

	events = append(events, secondItem)

	thirdBody, err := json.Marshal(BodyStructureOne{Amount: 350, Name: "George"})
	Utils.FailOnError(err, "error on create body")

	thirdItem := new(Event.Event)
	thirdItem.Initialize("resource.custom.receive", bytes.NewBuffer(thirdBody).String())

	events = append(events, thirdItem)

	fourthBody, err := json.Marshal(BodyStructureOne{Amount: 550, Name: "James"})
	Utils.FailOnError(err, "error on create body")

	fourthItem := new(Event.Event)
	fourthItem.Initialize("resource.custom.receive", bytes.NewBuffer(fourthBody).String())

	events = append(events, fourthItem)

	singleBody, err := json.Marshal(BodyStructureOne{Amount: 30, Name: "Willian"})
	Utils.FailOnError(err, "error on create body")

	singleItem := new(Event.Event)
	singleItem.Initialize("resource.custom.private.service", bytes.NewBuffer(singleBody).String())

	events = append(events, singleItem)

	actionBody, err := json.Marshal(BodyStructureOne{Amount: 30, Name: "Willian"})
	Utils.FailOnError(err, "error on create body")

	actionItem := new(Event.Event)
	actionItem.Initialize("resource.origin.action", bytes.NewBuffer(actionBody).String())

	events = append(events, actionItem)

	new(Models.Emitter).Trigger(events, *rabbit)

	<-forever
}

func main() {
	RunEmitter()
}
