package EventPeople

func ListenTo(eventName string, callback Callback) {
	Config.Broker.Consume(eventName, callback)
}
