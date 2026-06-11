package EventPeople

import "encoding/json"

type ContextDelivery struct {
	delivery    *DeliveryStruct
	callback    Callback
	rabbitCtx   *RabbitContext
}

type Job struct {
	job ContextDelivery
}

/*implement work interface*/
func (j *Job) Do() {
	var eventMessage Event
	json.Unmarshal(j.job.delivery.Body, &eventMessage)

	eventMessage.Name = j.job.delivery.RoutingKey
	eventMessage.SchemaVersion = eventMessage.Headers.SchemaVersion
	eventMessage.RetryCount = j.job.delivery.RetryCount

	// Prefer the full RabbitContext (which carries the channel for retry publishing)
	// if it was provided; otherwise fall back to constructing a minimal context.
	var ctx ContextInterface
	if j.job.rabbitCtx != nil {
		ctx = j.job.rabbitCtx
	} else {
		rabbitContext := NewContext(j.job.delivery.DeliveryInterface)
		rabbitContext.DeliveryStruct = *j.job.delivery
		ctx = rabbitContext
	}

	j.job.callback(eventMessage, ctx)
}
