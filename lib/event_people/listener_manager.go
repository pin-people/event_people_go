package EventPeople

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/pin-people/event_people_go/lib/event_people/worker"
)

type ListenerMethod func(event Event, ctx ContextInterface)

type ListenerManagerStruct struct {
	EventName string
	Method    ListenerMethod
	Listener  *BaseListener
}

type manager struct{}

type ListenerManagerContext struct {
	ListenerManager ListenerManagerStruct
	Event           Event
}

var ListenerManager = new(manager)

var ListenerConfigurationsList []ListenerManagerStruct

var Pool *worker.Pool

func (manager manager) BindAllListeners() {
	for index := range ListenerConfigurationsList {
		listenerItem := ListenerConfigurationsList[index]
		fmt.Println("Bind Event EventName =>", listenerItem.EventName)
		SubscribeTo(listenerItem.EventName)
	}
}

func (manager manager) ConsumeAllListeners() {
	workerPool, _ := strconv.Atoi(os.Getenv("WORKERS"))
	if workerPool == 0 {
		workerPool = runtime.NumCPU() * 2
	}
	fmt.Println("Starting worker pool with", workerPool, "workers")
	Pool = worker.NewWorkerPool(workerPool)
	Pool.Start()

	index := 0
	maxIndex := len(ListenerConfigurationsList) - 1

	for {
		if !Pool.IsWorkerAvailable() {
			time.Sleep(1 * time.Second)
			continue
		}

		if index > maxIndex {
			index = 0
		}

		listenerItem := ListenerConfigurationsList[index]
		delivery := GetMessage(listenerItem.EventName)

		if delivery == nil || len(delivery.Body) == 0 {
			index++
			continue
		}

		Pool.Submit(&Job{
			job: ContextDelivery{delivery, func(event Event, contextEvent ContextInterface) {
				listenerItem.Method(event, contextEvent)
			}},
		})
		Pool.AddWorkerCount()
		index++
	}
}

func (manager *manager) Register(model ListenerManagerStruct) {
	ListenerConfigurationsList = append(ListenerConfigurationsList, model)
}
