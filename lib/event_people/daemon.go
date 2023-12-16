package EventPeople

func DaemonStart() error {
	var forever chan struct{}
	ListenerManager.BindAllListeners()
	ListenerManager.ConsumeAllListeners()
	defer DaemonStop()
	<-forever
	return nil
}

func DaemonStop() {
	Config.Broker.CloseConnection()
}
