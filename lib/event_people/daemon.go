package EventPeople

type Daemon struct{}

func (d *Daemon) Start() {
	var forever chan struct{}
	ListenerManager.BindAllListeners()
	defer Config.Broker.CloseConnection()
	<-forever
}

func (d *Daemon) Stop() {
	Config.Broker.CloseConnection()
}
