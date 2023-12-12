package EventPeople

type RabbitContext struct {
	ContextInterface
	delivery DeliveryInterface
}

func (context *RabbitContext) Initialize(delivery DeliveryInterface) {
	context.delivery = delivery
}

func (context *RabbitContext) Success() {
	context.delivery.Ack(false)
}

func (context *RabbitContext) Fail() {
	context.delivery.Nack(false, false)
}

func (context *RabbitContext) Reject() {
	context.delivery.Reject(false)
}

func NewContext(delivery DeliveryInterface) *RabbitContext {
	context := &RabbitContext{delivery: delivery}
	return context
}
