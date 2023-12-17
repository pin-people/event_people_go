package EventPeople

import (
	"fmt"
	"log"
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
var workerPool int

func (manager manager) BindAllListeners() {
	for index := range ListenerConfigurationsList {
		listenerItem := ListenerConfigurationsList[index]
		err := SubscribeTo(listenerItem.EventName)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (manager manager) ConsumeAllListeners() {
	workerPool, _ = strconv.Atoi(os.Getenv("WORKERS"))
	if workerPool == 0 {
		workerPool = runtime.NumCPU()
	}
	Pool = worker.NewWorkerPool(workerPool)
	Pool.Start()

	for index := range ListenerConfigurationsList {
		listenerItem := ListenerConfigurationsList[index]
		go ListenTo(listenerItem.EventName, func(event Event, context ContextInterface) {
			for {
				if !Pool.IsWorkerAvailable() {
					time.Sleep(2 * time.Second)
					continue
				}
				break
			}
			delivery := context.(*RabbitContext)
			if delivery.DeliveryStruct.DeliveryInterface != nil {
				go Pool.Submit(&Job{
					job: ContextDelivery{&delivery.DeliveryStruct, func(event Event, contextEvent ContextInterface) {
						listenerItem.Method(event, contextEvent)
					}},
				})
				return
			}
		})
	}
}

func (manager *manager) Register(model ListenerManagerStruct) {
	ListenerConfigurationsList = append(ListenerConfigurationsList, model)
}

func printWorkerStatus() {
	var iterator = 0
	for {
		time.Sleep(100 * time.Millisecond)
		workersInUse, maxWorkersInPool := Pool.GetWorkerStatus()
		fmt.Printf("\r %d workers in use of %d status [%s]", workersInUse, maxWorkersInPool, getProgress(iterator))
		iterator++
	}
}

func getProgress(iteration int) string {
	indicators := []string{"-", "\\", "|", "/"}
	return indicators[iteration%len(indicators)]
}
