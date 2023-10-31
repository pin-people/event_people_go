package EventPeople

import "encoding/json"

func GetMessage(eventName string) *DeliveryStruct {
	return Config.Broker.Consume(eventName)
}

func GetMessageWithCallback(eventName string, callback Callback) {
	delivery := Config.Broker.Consume(eventName)
	if delivery != nil && len(delivery.Body) == 0 {
		var eventMessage Event
		json.Unmarshal(delivery.Body, &eventMessage)

		eventMessage.Name = eventMessage.Headers.AppName
		eventMessage.SchemaVersion = eventMessage.Headers.SchemaVersion

		rabbitContext := NewContext(delivery.DeliveryInterface)
		callback(eventMessage, rabbitContext)
	}
}

func SubscribeTo(eventName string) {
	Config.Broker.Subscribe(eventName)
}
