package EventPeople

func DaemonStart() {
	var forever chan struct{}
	ListenerManager.BindAllListeners()
	defer Config.Broker.CloseConnection()
	<-forever
}

func DaemonStop() {
	Config.Broker.CloseConnection()
}
