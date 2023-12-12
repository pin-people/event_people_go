package EventPeople

type configStruct struct {
	Broker AbstractBaseBroker
	UseDLX bool
}

var Config = new(configStruct)

func (config *configStruct) Init() {
	Config.Broker = new(RabbitBroker)
	Config.UseDLX = false
	Config.Broker.Init()
}

func (config *configStruct) CloseConnection() {
	Config.Broker.CloseConnection()
}
