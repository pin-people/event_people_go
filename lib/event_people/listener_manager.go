package EventPeople

type ListenerMethod func(event Event)

type ListenerManagerStruct struct {
    EventName string
    Method    ListenerMethod
    Listener  *BaseListener
}

type manager struct{}

var ListenerManager = new(manager)

var ListenerConfigurationsList []ListenerManagerStruct

func (manager manager) BindAllListeners() {
    for index := range ListenerConfigurationsList {
        listenerItem := ListenerConfigurationsList[index]
        ListenTo(listenerItem.EventName, func(event Event, context ContextInterface) {
            listenerItem.Listener.Initialize(context)
            listenerItem.Method(event)
        })
    }
}

func (manager *manager) Register(model ListenerManagerStruct) {
    ListenerConfigurationsList = append(ListenerConfigurationsList, model)
}
