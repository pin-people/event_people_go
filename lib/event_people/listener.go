// Package EventPeople is used to run lib
package EventPeople

func ListenTo(eventName string, callback Callback) {
	Config.Broker.Consume(eventName, callback)
}

func SubscribeTo(eventName string) {
	Config.Broker.Subscribe(eventName)
}
