package EventPeople

import (
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ContextDelivery struct {
	delivery       *DeliveryStruct
	callback       Callback
	retryQueueName string
	amqpChannel    retryPublisher
	retryConfig    RetryConfig
}

type Job struct {
	job ContextDelivery
}

// Do implements the worker.Work interface. It decodes the delivery, builds a
// retry-aware RabbitContext, and dispatches to the user callback.
func (j *Job) Do() {
	var eventMessage Event
	json.Unmarshal(j.job.delivery.Body, &eventMessage)

	eventMessage.Name = j.job.delivery.RoutingKey
	eventMessage.SchemaVersion = eventMessage.Headers.SchemaVersion

	// Read x-event-people-retries from the delivery headers if available.
	if amqpDelivery, ok := j.job.delivery.DeliveryInterface.(amqp.Delivery); ok {
		if amqpDelivery.Headers != nil {
			if retryVal, ok := amqpDelivery.Headers["x-event-people-retries"]; ok {
				switch v := retryVal.(type) {
				case int32:
					eventMessage.RetryCount = int(v)
				case int64:
					eventMessage.RetryCount = int(v)
				case int:
					eventMessage.RetryCount = v
				}
			}
		}
	}

	retryConfig := j.job.retryConfig
	if retryConfig.MaxAttempts == 0 {
		retryConfig = Config.GetRetryConfig()
	}

	var rabbitContext *RabbitContext
	if j.job.amqpChannel != nil && j.job.retryQueueName != "" {
		rabbitContext = NewContextWithRetry(
			j.job.delivery.DeliveryInterface,
			j.job.amqpChannel,
			j.job.retryQueueName,
			eventMessage.RetryCount,
			retryConfig.MaxAttempts,
			retryConfig.DLQName,
			retryConfig.InitialDelay,
			retryConfig.DelayStrategy,
		)
	} else {
		rabbitContext = NewContext(j.job.delivery.DeliveryInterface)
	}
	rabbitContext.DeliveryStruct = *j.job.delivery

	j.job.callback(eventMessage, rabbitContext)
}
