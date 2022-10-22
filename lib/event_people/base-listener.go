package EventPeople

import (
	"os"
	"strings"
)

type AbstractBaseListener interface {
	Initialize(context ContextInterface)
	Success()
	Fail()
	Reject()
}

type BaseListener struct {
	AbstractBaseListener
	context ContextInterface
}

func (base *BaseListener) Initialize(context ContextInterface) {
	base.context = context
}

func (base *BaseListener) Success() {
	base.context.Success()
}

func (base *BaseListener) Fail() {
	base.context.Fail()
}

func (base *BaseListener) Reject() {
	base.context.Reject()
}

func (base *BaseListener) BindEvent(method ListenerMethod, eventName string) {
	eventNameSplited := strings.Split(eventName, ".")

	if len(eventNameSplited) <= 4 {
		ListenerManager.Register(
			ListenerManagerStruct{
				EventName: FixedEventName(eventName, "all"),
				Method:    method,
				Listener:  base,
			},
		)
		ListenerManager.Register(
			ListenerManagerStruct{
				EventName: FixedEventName(eventName, os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME")),
				Method:    method,
				Listener:  base,
			},
		)
		return
	}
	ListenerManager.Register(
		ListenerManagerStruct{
			EventName: FixedEventName(eventName, os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME")),
			Method:    method,
			Listener:  base,
		},
	)
}

func FixedEventName(eventName string, postfix string) string {
	eventNameSplited := strings.Split(eventName, ".")

	if len(eventNameSplited) == 4 {
		eventNameSplited[3] = postfix
		return strings.Join(eventNameSplited, ".")
	}
	eventNameSplited = append(eventNameSplited, postfix)
	return strings.Join(eventNameSplited, ".")
}
