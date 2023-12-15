package EventPeople

type configStruct struct {
	Broker AbstractBaseBroker
	UseDLX bool
}

var Config = configStruct{Broker: nil, UseDLX: false}

func (config *configStruct) Init() {
	Config.Broker = new(RabbitBroker)
	Config.Broker.Init()
}

func (config *configStruct) CloseConnection() {
	Config.Broker.CloseConnection()
}
