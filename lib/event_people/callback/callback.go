package Callback

import (
	ContextEvent "github.com/pinpeople/event_people_go/lib/event_people/context-event"
	Event "github.com/pinpeople/event_people_go/lib/event_people/event"
)

type Callback func(event Event.Event, listener ContextEvent.BaseContextEvent)
