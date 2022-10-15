package EventPeople

type configStruct struct {
	Broker AbstractBaseBroker
}

var Config = new(configStruct)

func (config *configStruct) Init() {
	Config.Broker = new(RabbitBroker)
	Config.Broker.Init()
}

func (config *configStruct) CloseConnection() {
	Config.Broker.CloseConnection()
}
