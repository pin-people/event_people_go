# EventPeople

[![CircleCI](https://dl.circleci.com/status-badge/img/gh/pin-people/event_people_node/tree/main.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/pin-people/event_people_node/tree/main)

EventPeople is a tool to simplify the communication of event based services. It is an based on the [EventBus](https://github.com/EmpregoLigado/event_bus_rb) gem.

The main idea is to provide a tool that can emit or consume events based on its names, the event name has 4 words (`resource.origin.action.destination`) which defines some important info about what kind of event it is, where it comes from and who is eligible to consume it:

-   **resource:** Defines which resource this event is related like a `user`, a `product`, `company` or anything that you want;
-   **origin:** Defines the name of the system which emitted the event;
-   **action:** What action is made on the resource like `create`, `delete`, `update`, etc. PS: _It is recommended to use the Semple Present tense for actions_;
-   **destination (Optional):** This word is optional and if not provided EventPeople will add a `.all` to the end of the event name. It defines which service should consume the event being emitted, so if it is defined and there is a service whith the given name only this service will receive it. It is very helpful when you need to re-emit some events. Also if it is `.all` all services will receive it.

As of today EventPeople uses RabbitMQ as its datasource, but there are plans to add support for other Brokers in the future.

## Installation

Add this line to your application's `go.mod`:

```yaml
require github.com/pin-people/event_people_go
```

To install and add it as a dependency in your project:

```cmd
   $ go get "github.com/pin-people/event_people_go"
```

You need to install in mode proxy, to this use:

### Linux/MacOS

```cmd
    $ GOPROXY=https://proxy.golang.org GO111MODULE=on go get github.com/pin-people/event_people_go
```

### Windows

```cmd
    $ set GOPROXY=https://proxy.golang.org; set GO111MODULE=on; go get github.com/pin-people/event_people_go
```

Set env vars and execute init function:

```golang
func init() {
	os.Setenv("WORKERS", "4")
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "service")
	os.Setenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME", "event_people")
	os.Setenv("RABBIT_EVENT_PEOPLE_VHOST", "event_people")
	os.Setenv("RABBIT_URL", "amqp://admin:admin@localhost:5672")

	EventPeople.Config.Init()
}
```

## Usage

### Events

The main component of `EventPeople` is the `EventPeople.Event` class which wraps all the logic of an event and whenever you receive or want to send an event you will use it.

It has 2 attributes `name` and `payload`:

-   **name:** The name must follow our conventions, being it 3 (`resource.origin.action`) or 4 words (`resource.origin.action.destination`);
-   **payload:** It is the body of the massage, it should be a Custom Struct mapping the JSON fields of the message.

```golang
import (
  EventPeople "github.com/pin-people/event_people_go"
)

type BodyStructure struct {
	Amount int    `json:"amount"`
	Name   string `json:"name"`
}

func main() {
  var eventName = "user.users.create";
  var body = BodyStructure{ id: 42, name: "John Doe", age: 35 };
  var event = EventPeople.NewEvent(event_name, body);
}

```

There are 3 main interfaces to use `EventPeople` on your project:

-   Calling `EventPeople.TriggerEmitter(event []*EventPeople.Event)` inside your project;
-   Calling `EventPeople.ListenTo(eventName string)` inside your project;
-   Or extending `EventPeople.BaseListener` and use it as a daemon.

### Using the Emitter

You can emit events on your project passing an `EventPeople.Event` instance to the `EventPeople.TriggerEmitter` method. Doing this other services that are subscribed to these events will receive it.

```golang
import (
  "encoding/json"
  EventPeople "github.com/pin-people/event_people_go"
)

type BodyStructureEmmiter struct {
	Amount int    `json:"amount"`
	Name   string `json:"name"`
}

func main() {
  var eventName = "receipt.payments.pay.users"
  var body := BodyStructureEmmiter{Amount: 350.76, Name: "John"}


  event := EventPeople.NewEvent(eventName, body)

  EventPeople.TriggerEmitter([]*EventPeople.Event{event})

  // Don't forget to close the connection!!!
  EventPeople.Config.CloseConnection()
}

```

[See more details](https://github.com/pin-people/event_people_node/blob/master/examples/emitter.rb)

### Listeners

You can subscribe to events based on patterns for the event names you want to consume or you can use the full name of the event to consume single events.

We follow the RabbitMQ pattern matching model, so given each word of the event name is separated by a dot (`.`), you can use the following symbols:

-   `* (star):` to match exactly one word. Example `resource.*.*.all`;
-   `# (hash):` to match zero or more words. Example `resource.#.all`.

Other important aspect of event consumming is the result of the processing we provide 3 methods so you can inform the Broker what to do with the event next:

-   `Success:` should be called when the event was processed successfuly and the can be discarded;
-   `Fail:` should be called when an error ocurred processing the event and the message should be requeued;
-   `Reject:` should be called whenever a message should be discarded without being processed.

Given you want to consume a single event inside your project you can use the `EventPeople.ListenTo` method. It consumes a single event, given there are events available to be consumed with the given name pattern.

```golang
import (
  "fmt"
  EventPeople "github.com/pin-people/event_people_go"
)

func main() {
  // 3 words event names will be replaced by its 4 word wildcard
  // counterpart: 'payment.payments.pay.all'
  var eventName = "payment.payments.pay"
  var once = make(chan int)

  EventPeople.ListenTo(eventName, func (event EventPeople.Event, context EventPeople.BaseListener) {
    msg := event.Body

		fmt.Println("")
		fmt.Println(fmt.Sprintf("  - Received the %s message from %s:", event.Name, event.Headers.Origin))
		fmt.Println(fmt.Sprintf("     Message: %s", msg))
		fmt.Println("")
		context.Success()
    once <- 1
  });
  <-once
  EventPeople.Config.CloseConnection()
}
```

You can also receive all available messages using a channel and time sleep:

```golang
import (
  "fmt"
  EventPeople "github.com/pin-people/event_people_go"
)
var once = make(chan int)

func main() {
  var eventName = "payment.payments.pay.all"

	EventPeople.ListenTo(eventName, func(event EventPeople.Event, context EventPeople.BaseListener) {
		msg := event.Body

		fmt.Println("")
		fmt.Println(fmt.Sprintf("  - Received the %s message from %s:", event.Name, event.Headers.Origin))
		fmt.Println(fmt.Sprintf("     Message: %s", msg))
		fmt.Println("")
		context.Success()
	})

	go func() {
    time.Sleep(15 * time.Second)
		once <- 1
	}()

	<-once
  EventPeople.Config.CloseConnection()
}
```

[See more details](https://github.com/pin-people/event_people_node/blob/master/examples/listener.rb)

#### Multiple events routing

If your project needs to handle lots of events you can extend `EventPeople.BaseListener` class to bind how many events you need to instance methods, so whenever an event is received the method will be called automatically.

```golang
import (
	"encoding/json"
	"fmt"
	"os"

	EventPeople "github.com/pin-people/event_people_go/lib/event_people"
)

func init() {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "service")
	os.Setenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME", "event_people")
	os.Setenv("RABBIT_EVENT_PEOPLE_VHOST", "event_people")
	os.Setenv("RABBIT_URL", "amqp://admin:admin@localhost:5672")
	os.Setenv("RABBIT_FULL_URL", fmt.Sprintf("%s/%s", os.Getenv("RABBIT_URL"), os.Getenv("RABBIT_EVENT_PEOPLE_VHOST")))

	EventPeople.Config.Init()
}

type BodyStructureDaemon struct {
	Amount int    `json:"amount"`
	Name   string `json:"name"`
}

type PrivateMessageDaemon struct {
	Message string `json:"message"`
}

type SecondPrivateMessageDaemon struct {
	Bo string `json:"bo"`
	Dy string `json:"dy"`
}

func pay(event EventPeople.Event, cel EventPeople.ContextInterface) {
	var bodyDaemon = new(BodyStructureDaemon)
	event.SetStructBody(&bodyDaemon)

	fmt.Println(fmt.Sprintf("Paid %v for %s ~> %s", bodyDaemon.Amount, bodyDaemon.Name, event.Name))
	cel.Success()
}

func receive(event EventPeople.Event, cel EventPeople.ContextInterface) {
	var bodyDaemon = new(BodyStructureDaemon)
  event.SetStructBody(&bodyDaemon)

	if bodyDaemon.Amount < 500 {
		fmt.Println(fmt.Sprintf("[consumer] Got SKIPPED message:\n%d from %s ~> %s", bodyDaemon.Amount, bodyDaemon.Name, event.Name))
		cel.Reject()
		return
	}
	fmt.Println("Received %d from %s ~> %s", bodyDaemon.Amount, bodyDaemon.Name, event.Name)
	cel.Success()
}

func privateChannel(event EventPeople.Event, cel EventPeople.ContextInterface) {
	var bodyDaemon = new(PrivateMessageDaemon)
  event.SetStructBody(&bodyDaemon)

	fmt.Println(fmt.Sprintf("[Consumer] Got a private message: %s ~> %s", bodyDaemon.Message, event.Name))
	cel.Success()
}

func main() {
	EventPeople.BindEvent(pay, "resource.custom.pay")
	EventPeople.BindEvent(receive, "resource.custom.receive")
	EventPeople.BindEvent(privateChannel, "resource.custom.private.service")
	EventPeople.BindEvent(secondPrivateChannel, "resource.origin.action.service")

	EventPeople.DaemonStart()
}
```

[See more details](https://github.com/pin-people/event_people_node/blob/master/examples/daemon.rb)

#### Creating a Daemon

If you have the need to create a deamon to consume messages on background you can use the `EventPeople.DaemonStart` method to do so with ease. Just remember to define or import all the event bindings before starting the daemon.

```golang
import (
	"encoding/json"
	"fmt"
	"os"

	EventPeople "github.com/pin-people/event_people_go/lib/event_people"
)

func init() {
	os.Setenv("RABBIT_EVENT_PEOPLE_APP_NAME", "service")
	os.Setenv("RABBIT_EVENT_PEOPLE_TOPIC_NAME", "event_people")
	os.Setenv("RABBIT_EVENT_PEOPLE_VHOST", "event_people")
	os.Setenv("RABBIT_URL", "amqp://admin:admin@localhost:5672")
	os.Setenv("RABBIT_FULL_URL", fmt.Sprintf("%s/%s", os.Getenv("RABBIT_URL"), os.Getenv("RABBIT_EVENT_PEOPLE_VHOST")))

	EventPeople.Config.Init()
}

type BodyStructureDaemon struct {
	Amount int    `json:"amount"`
	Name   string `json:"name"`
}

type PrivateMessageDaemon struct {
	Message string `json:"message"`
}

type SecondPrivateMessageDaemon struct {
	Bo string `json:"bo"`
	Dy string `json:"dy"`
}

func pay(event EventPeople.Event, cel EventPeople.BaseListener) {
	var bodyDaemon = new(BodyStructureDaemon)
	event.SetStructBody(&bodyDaemon)

	fmt.Println(fmt.Sprintf("Paid %v for %s ~> %s", bodyDaemon.Amount, bodyDaemon.Name, event.Name))
	cel.Success()
}

func receive(event EventPeople.Event, cel EventPeople.BaseListener) {
	var bodyDaemon = new(BodyStructureDaemon)
	event.SetStructBody(&bodyDaemon)

	if bodyDaemon.Amount < 500 {
		fmt.Println(fmt.Sprintf("[consumer] Got SKIPPED message:\n%d from %s ~> %s", bodyDaemon.Amount, bodyDaemon.Name, event.Name))
		cel.Reject()
		return
	}
	fmt.Println("Received %d from %s ~> %s", bodyDaemon.Amount, bodyDaemon.Name, event.Name)
	cel.Success()
}

func privateChannel(event EventPeople.Event, cel EventPeople.BaseListener) {
	var bodyDaemon = new(PrivateMessageDaemon)
	event.SetStructBody(&bodyDaemon)

	fmt.Println(fmt.Sprintf("[Consumer] Got a private message: %s ~> %s", bodyDaemon.Message, event.Name))
	cel.Success()
}

func main() {
  	EventPeople.BindEvent(pay, "resource.custom.pay")
	EventPeople.BindEvent(receive, "resource.custom.receive")
	EventPeople.BindEvent(privateChannel, "resource.custom.private.service")
	EventPeople.BindEvent(secondPrivateChannel, "resource.origin.action.service")

	EventPeople.DaemonStart()
}
```

[See more details](https://github.com/pin-people/event_people_node/blob/master/examples/daemon.rb)

## Development

To install this module onto your local machine, run `go get`.

## Contributing

-   Fork it
-   Create your feature branch (`git checkout -b my-new-feature`)
-   Commit your changes (`git commit -am 'Add some feature'`)
-   Push to the branch (`git push origin my-new-feature`)
-   Create a new Pull Request

## License

The module is available as open source under the terms of the [LGPL 3.0 License](https://www.gnu.org/licenses/lgpl-3.0.en.html).
