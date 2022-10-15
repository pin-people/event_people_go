package EventPeople

import (
	"os"
	"strings"
)

type Headers struct {
	AppName       string  `json:"appName"`
	Resource      string  `json:"resource"`
	Origin        string  `json:"origin"`
	Action        string  `json:"action"`
	Destination   string  `json:"destination"`
	SchemaVersion float64 `json:"schemaVersion"`
}

type Event struct {
	Name          string  `json:"name"`
	Headers       Headers `json:"headers"`
	Body          any     `json:"body"`
	SchemaVersion float64 `json:"schemaVersion"`
}

type Payload struct {
	Headers Headers `json:"headers"`
	Body    any     `json:"body"`
}

func NewEvent(name string, body any, schemaVersion ...float64) *Event {
	event := new(Event)
	event.Initialize(name, body, schemaVersion...)
	return event
}

func (event *Event) Initialize(name string, body any, schemaVersion ...float64) {
	event.Name = name
	event.SchemaVersion = 1.0
	event.Body = StructToJsonString(body)

	if schemaVersion != nil {
		event.SchemaVersion = schemaVersion[0]
	}

	if name != "" {
		event.generateHeaders()
		event.fixName()
	}
}

func (event *Event) Payload() string {
	payload := Payload{
		event.Headers, event.Body,
	}
	return StructToJsonString(payload)
}

func (event *Event) HasBody() bool {
	return event.Body != nil
}

func (event *Event) HasName() bool {
	return event.Name != ""
}

func (event *Event) generateHeaders() {
	headerSpec := strings.Split(event.Name, ".")

	if len(headerSpec) == 3 {
		headerSpec = append(headerSpec, "all")
	}

	event.Headers = Headers{
		AppName:       os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME"),
		Resource:      headerSpec[0],
		Origin:        headerSpec[1],
		Action:        headerSpec[2],
		Destination:   headerSpec[3],
		SchemaVersion: event.SchemaVersion,
	}
}

func (event *Event) fixName() {
	headerSpec := strings.Split(event.Name, ".")

	if len(headerSpec) == 3 {
		headerSpec = append(headerSpec, "all")
		name := strings.Join(headerSpec, ".")
		event.Name = event.Headers.AppName + "-" + name
	} else {
		event.Name = event.Headers.AppName + "-" + event.Name
	}
}

func (event *Event) GetRoutingKey() string {
	return event.Headers.AppName + "-" + event.Headers.Resource + "." + event.Headers.Origin + "." + event.Headers.Action + "." + event.Headers.Destination
}
