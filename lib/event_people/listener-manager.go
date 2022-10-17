package EventPeople

type ManagerMethod func(event Event)

type ListenerManagerStruct struct {
	RoutingKey string
	Method     ManagerMethod
	Listener   *BaseListener
}

type manager struct{}

var ListenerManager = new(manager)

var ListenerConfigurationsList []ListenerManagerStruct

func (manager manager) BindAllListeners() {
	for index := range ListenerConfigurationsList {
		listenerItem := ListenerConfigurationsList[index]
		NewListener().On(listenerItem.RoutingKey, func(event Event, listener BaseListener) {
			listenerItem.Listener.Initialize(listener.context, listener.DeliveryInfo)
			listenerItem.Method(event)
		})
	}
}

func (manager *manager) Register(model ListenerManagerStruct) {
	ListenerConfigurationsList = append(ListenerConfigurationsList, model)
}
