// Package EventPeople is used to run lib
package EventPeople

func ListenTo(eventName string, callback Callback) {
	Config.Broker.Consume(eventName, callback)
}

func SubscribeTo(eventName string) error {
	return Config.Broker.Subscribe(eventName)
}
