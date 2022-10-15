package EventPeople

import (
	"os"
	"strings"
)

type AbstractBaseListener interface {
	Initialize(context ContextInterface, deliveryInfo DeliveryInfo)
	Callback(methodName string, event any)
	Success()
	Fail()
	Reject()
}

type BaseListener struct {
	AbstractBaseListener
	context      ContextInterface
	DeliveryInfo DeliveryInfo
}

func (base *BaseListener) Initialize(context ContextInterface, deliveryInfo DeliveryInfo) {
	base.context = context
	base.DeliveryInfo = deliveryInfo
}

func (base *BaseListener) Success() {
	base.context.Ack(false)
}

func (base *BaseListener) Fail() {
	base.context.Nack(false, true)
}

func (base *BaseListener) Reject() {
	base.context.Reject(false)
}

func (base *BaseListener) BindEvent(method ManagerMethod, eventName string) {
	ListenerManager.Register(
		ListenerManagerStruct{
			RoutingKey: base.fixedEventName(eventName, "all"),
			Method:     method,
			Listener:   base,
		},
	)
	ListenerManager.Register(
		ListenerManagerStruct{
			RoutingKey: base.fixedEventName(eventName, os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME")),
			Method:     method,
			Listener:   base,
		},
	)
}

func (base *BaseListener) fixedEventName(eventName string, postfix string) string {
	eventNameSplited := strings.Split(eventName, ".")

	if len(eventNameSplited) == 4 {
		eventNameSplited[3] = postfix
		return strings.Join(eventNameSplited, ".")
	}
	eventNameSplited = append(eventNameSplited, postfix)
	return strings.Join(eventNameSplited, ".")
}
