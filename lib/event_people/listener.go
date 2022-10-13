package EventPeople

type ListenerInterface interface {
	On(eventName string, callback Callback)
}

type Listener struct{}

func (listener *Listener) On(eventName string, callback Callback) {
	Config.Broker.Consume(eventName, callback)
}
