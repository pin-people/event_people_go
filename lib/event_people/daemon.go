package EventPeople

import (
	"os"
	"os/signal"
	"syscall"
)

func DaemonStart() error {
	ListenerManager.BindAllListeners()
	ListenerManager.ConsumeAllListeners()
	defer DaemonStop()
	bindSignals()
	return nil
}

func DaemonStop() {
	Config.Broker.CloseConnection()
}

func bindSignals() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
}
