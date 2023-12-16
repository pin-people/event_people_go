package EventPeople

import "encoding/json"

type ContextDelivery struct {
	delivery *DeliveryStruct
	callback Callback
}

type Job struct {
	job ContextDelivery
}

/*implement work interface*/
func (j *Job) Do() {
	var eventMessage Event
	json.Unmarshal(j.job.delivery.Body, &eventMessage)

	eventMessage.Name = eventMessage.Headers.AppName
	eventMessage.SchemaVersion = eventMessage.Headers.SchemaVersion

	rabbitContext := NewContext(j.job.delivery.DeliveryInterface)
	j.job.callback(eventMessage, rabbitContext)
}
