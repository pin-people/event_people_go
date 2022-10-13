package EventPeople

type AbstractBaseBroker interface {
	Init()
	GetConnection() any
	GetConsumers() int
	Channel()
	Consume(eventName string, callback Callback)
	Produce(events Event)
	RabbitURL() string
	CloseConnection()
}

type BaseBroker struct {
	AbstractBaseBroker
}
