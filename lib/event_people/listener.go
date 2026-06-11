// Package EventPeople is used to run lib
package EventPeople

// ListenTo subscribes to eventName and invokes callback for each delivery.
// An optional RetryConfig may be supplied to override the defaults from Config.
func ListenTo(eventName string, callback Callback, retryConfig ...RetryConfig) {
	Config.Broker.Consume(eventName, callback, retryConfig...)
}

func SubscribeTo(eventName string) error {
	return Config.Broker.Subscribe(eventName)
}
