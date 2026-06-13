package EventPeople

// ContextInterface is passed to every user callback. It informs the broker
// of the processing outcome and exposes retry metadata.
type ContextInterface interface {
	Initialize(DeliveryInterface)
	Success()
	Fail()
	Reject()
	GetMaxRetries() int
	GetIsLastRetry() bool
}

type DeliveryInterface interface {
	Ack(bool) error
	Nack(bool, bool) error
	Reject(bool) error
}

type DeliveryStruct struct {
	DeliveryInterface
	DeliveryTag   uint64
	Body          []byte
	RoutingKey    string
	Headers       map[string]interface{}
	ContentType   string
	MaxRetries    int
	DelayStrategy string
	QueueName     string
	RetryCount    int
}
