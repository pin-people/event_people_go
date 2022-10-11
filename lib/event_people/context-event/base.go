package ContextEvent

import (
	Context "github.com/pinpeople/event_people_go/lib/event_people/context"
	DeliveryInfo "github.com/pinpeople/event_people_go/lib/event_people/delivery-info"
)

type AbstractContextEvent interface {
	Initialize(context Context.ContextInterface, deliveryInfo DeliveryInfo.DeliveryInfo)
	Callback(methodName string, event any)
	Success()
	Fail()
	Reject()
}

type BaseContextEvent struct {
	AbstractContextEvent
	context      Context.ContextInterface
	DeliveryInfo DeliveryInfo.DeliveryInfo
}

func (base *BaseContextEvent) Initialize(context Context.ContextInterface, deliveryInfo DeliveryInfo.DeliveryInfo) {
	base.context = context
	base.DeliveryInfo = deliveryInfo
}

func (base *BaseContextEvent) Success() {
	base.context.Ack(false)
}

func (base *BaseContextEvent) Fail() {
	base.context.Nack(false, true)
}

func (base *BaseContextEvent) Reject() {
	base.context.Reject(false)
}
