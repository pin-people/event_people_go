package EventPeople

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/pin-people/event_people_go/lib/event_people/worker"
)

func DaemonStart() {
	var forever chan struct{}

	ListenerManager.BindAllListeners()

	workerPool, _ := strconv.Atoi(os.Getenv("WORKERS"))
	if workerPool == 0 {
		workerPool = runtime.NumCPU() * 2
	}
	fmt.Println("Starting worker pool with", workerPool, "workers")
	Pool = worker.NewWorkerPool(workerPool)
	Pool.Start()
	for i := 0; i < workerPool; i++ {
		go ListenerManager.ConsumeAllListeners()
	}
	defer Config.Broker.CloseConnection()
	<-forever
}

func DaemonStop() {
	Config.Broker.CloseConnection()
}
