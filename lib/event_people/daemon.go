package EventPeople

func DaemonStart() {
	var forever chan struct{}
	ListenerManager.BindAllListeners()
	ListenerManager.ConsumeAllListeners()
	defer Config.Broker.CloseConnection()
	<-forever
}

func DaemonStop() {
	Config.Broker.CloseConnection()
}
