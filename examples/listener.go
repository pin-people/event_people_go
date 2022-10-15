package main

import (
	"fmt"
	"os"
	"time"

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

func listenerCallback(event EventPeople.Event, context EventPeople.BaseListener) {
	msg := EventPeople.StructToJsonString(event.Body)
	fmt.Println(
		fmt.Sprintf("EventName: %s \nBody: %s", event.Name, msg),
	)
	context.Success()
	twice <- 1
}

var twice = make(chan int)

func RunListener() {
	EventPeople.NewListener().On("resource.custom.*", listenerCallback)
	EventPeople.NewListener().On("resource.origin.*", listenerCallback)
	<-twice
	<-twice
	EventPeople.Config.CloseConnection()
}

var once = make(chan int)

func main() {
	// RunListener()

	var eventName = "payment.payments.pay.all"

	EventPeople.NewListener().On(eventName, func(event EventPeople.Event, context EventPeople.BaseListener) {
		msg := EventPeople.StructToJsonString(event.Body)

		fmt.Println("")
		fmt.Println(fmt.Sprintf("  - Received the %s message from %s:", event.Name, event.Headers.Origin))
		fmt.Println(fmt.Sprintf("     Message: %s", msg))
		fmt.Println("")
		context.Success()
	})

	defer EventPeople.Config.CloseConnection()

	go func() {
		time.Sleep(15 * time.Second)
		once <- 1
	}()

	<-once
}
