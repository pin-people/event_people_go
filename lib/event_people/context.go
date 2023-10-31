package EventPeople

type ContextInterface interface {
	Initialize(DeliveryInterface)
	Success()
	Fail()
	Reject()
}

type DeliveryInterface interface {
	Ack(bool) error
	Nack(bool, bool) error
	Reject(bool) error
}

type DeliveryStruct struct {
	DeliveryInterface
	DeliveryTag uint64
	Body        []byte
}
