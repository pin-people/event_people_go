# EventPeople

[![CircleCI](https://dl.circleci.com/status-badge/img/gh/pin-people/event_people_node/tree/main.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/pin-people/event_people_node/tree/main)

EventPeople is a tool to simplify the communication of event based services. It is an based on the [EventBus](https://github.com/EmpregoLigado/event_bus_rb) gem.

The main idea is to provide a tool that can emit or consume events based on its names, the event name has 4 words (`resource.origin.action.destination`) which defines some important info about what kind of event it is, where it comes from and who is eligible to consume it:

- **resource:** Defines which resource this event is related like a `user`, a `product`, `company` or anything that you want;
- **origin:** Defines the name of the system which emitted the event;
- **action:** What action is made on the resource like `create`, `delete`, `update`, etc. PS: _It is recommended to use the Semple Present tense for actions_;
- **destination (Optional):** This word is optional and if not provided EventPeople will add a `.all` to the end of the event name. It defines which service should consume the event being emitted, so if it is defined and there is a service whith the given name only this service will receive it. It is very helpful when you need to re-emit some events. Also if it is `.all` all services will receive it.

As of today EventPeople uses RabbitMQ as its datasource, but there are plans to add support for other Brokers in the future.

## Installation

Add this line to your application's `go.mod`:

```yaml
require github.com/pinpeople/event-people-go
```

To install and add it as a dependency in your project:

    $ go get "github.com/pinpeople/event-people-go"

And set env vars:

```bash
export RABBIT_URL = 'amqp://guest:guest@localhost:5672'
export RABBIT_EVENT_PEOPLE_APP_NAME = 'service_name'
export RABBIT_EVENT_PEOPLE_VHOST = 'event_people'
export RABBIT_EVENT_PEOPLE_TOPIC_NAME = 'event_people'
```

## Usage

### Events

The main component of `EventPeople` is the `EventPeople.Event` class which wraps all the logic of an event and whenever you receive or want to send an event you will use it.

It has 2 attributes `name` and `payload`:

- **name:** The name must follow our conventions, being it 3 (`resource.origin.action`) or 4 words (`resource.origin.action.destination`);
- **payload:** It is the body of the massage, it should be a Hash object for simplicity and flexibility.

```golang
import (
  "github.com/pinpeople/event-people-go"
)

var event_name = "user.users.create";
var body = { id: 42, name: "John Doe", age: 35 };
var event = new EventPeople(event_name, body);
```

There are 3 main interfaces to use `EventPeople` on your project:

- Calling `EventPeople.Emitter.Trigger(event: Event)` inside your project;
- Calling `EventPeople.Listener.On(event_name: String)` inside your project;
- Or extending `EventPeople.BaseListeners` and use it as a daemon.

### Using the Emitter

You can emit events on your project passing an `EventPeople.Event` instance to the `EventPeople.Emitter.Trigger` method. Doing this other services that are subscribed to these events will receive it.

```golang
import (
  EventPeople "github.com/pinpeople/event-people-go"
)

const event_name = "receipt.payments.pay.users"
const body = { amount: 350.76 }
const event = new Event(event_name, body)

EventPeople.Emitter.Trigger(event)

// Don't forget to close the connection!!!
EventPeople.Config.close_connection()
```

[See more details](https://github.com/pin-people/event_people_node/blob/master/examples/emitter.rb)

### Listeners

You can subscribe to events based on patterns for the event names you want to consume or you can use the full name of the event to consume single events.

We follow the RabbitMQ pattern matching model, so given each word of the event name is separated by a dot (`.`), you can use the following symbols:

- `* (star):` to match exactly one word. Example `resource.*.*.all`;
- `# (hash):` to match zero or more words. Example `resource.#.all`.

Other important aspect of event consumming is the result of the processing we provide 3 methods so you can inform the Broker what to do with the event next:

- `success:` should be called when the event was processed successfuly and the can be discarded;
- `fail:` should be called when an error ocurred processing the event and the message should be requeued;
- `reject:` should be called whenever a message should be discarded without being processed.

Given you want to consume a single event inside your project you can use the `EventPeople.Listener.On` method. It consumes a single event, given there are events available to be consumed with the given name pattern.

```golang
import (
  "fmt"
  EventPeople "github.com/pinpeople/event-people-go"
)

// 3 words event names will be replaced by its 4 word wildcard
// counterpart: 'payment.payments.pay.all'
var event_name = "payment.payments.pay"

EventPeople.Listener.On(event_name, (event: Event, context: EventPeople.BaseListener) => {
  fmt.Println("")
  fmt.Println(`  - Received the "${event.name}" message from ${event.origin}:`);
  fmt.Println(`     Message: ${event.body}`);
  fmt.Println("")
  context.Success()
});

defer EventPeople.Config.Close_connection()
```

You can also receive all available messages using a loop:

```golang
import (
  "fmt"
  EventPeople "github.com/pinpeople/event-people-go"
)

var event_name = "payment.payments.pay.all";
var has_events = true;

while (has_events) {
  has_events = false;

  EventPeople.Listener.On("SOME_EVENT", (event: Event, context: EventPeople.Listener.
  ) => {
    has_events = true;
    fmt.Println("");
    fmt.Println(
      `  - Received the "${event.name}" message from ${event.origin}:`
    );
    fmt.Println(`     Message: ${event.body}`);
    fmt.Println("");
    context.success();
  });
}

EventPeople.Config.close_connection();
```

[See more details](https://github.com/pin-people/event_people_node/blob/master/examples/listener.rb)

#### Multiple events routing

If your project needs to handle lots of events you can extend `EventPeople.BaseListeners` class to bind how many events you need to instance methods, so whenever an event is received the method will be called automatically.

```golang
import (
  "fmt"
  EventPeople "github.com/pinpeople/event-people-go"
)

type CustomEventListener struct {
  EventPeople.BaseListeners
}

func (cel *CustomEventListener) RunListeners() {
  cel.bindEvent("resource.custom.pay", cel.pay);
  cel.bindEvent("resource.custom.receive", cel.receive);
  cel.bindEvent("resource.custom.private.service", cel.privateChannel);
}
func (cel *CustomEventListener) pay(event: Event) {
  fmt.Println(`Paid #{event.body['amount']} for #{event.body['name']} ~> #{event.name}`);

  this.success();
}
func (cel *CustomEventListener) receive(event: Event) {
  if (event.body.amount > 500) {
    fmt.Println(`Received ${event.body['amount']} from ${event.body['name']} ~> ${event.name}`);
  } else {
    fmt.Println("[consumer] Got SKIPPED message");
    return this.reject();
  }

  this.success();
}

func (cel *CustomEventListener) privateChannel(event: Event) {
  fmt.Println(`[consumer] Got a private message: "${event.body['message']}" ~> ${event.name}`);

  this.success();
}
```

[See more details](https://github.com/pin-people/event_people_node/blob/master/examples/daemon.rb)

#### Creating a Daemon

If you have the need to create a deamon to consume messages on background you can use the `EventPeople.Daemon.Start` method to do so with ease. Just remember to define or import all the event bindings before starting the daemon.

```golang
import (
  "fmt"
  EventPeople "github.com/pinpeople/event-people-go"
)

type CustomEventListener struct {
  EventPeople.BaseListeners
}

func (cel *CustomEventListener) RunListeners() {
  cel.bindEvent("resource.custom.pay", cel.pay);
  cel.bindEvent("resource.custom.receive", cel.receive);
  cel.bindEvent("resource.custom.private.service", cel.privateChannel);
}

func (cel *CustomEventListener) pay(event: Event) {
  fmt.Println(`Paid #{event.body['amount']} for #{event.body['name']} ~> #{event.name}`);

  this.success();
}
func (cel *CustomEventListener) receive(event: Event) {
  if (event.body.amount > 500) {
    fmt.Println(`Received ${event.body['amount']} from ${event.body['name']} ~> ${event.name}`);
  } else {
    fmt.Println("[consumer] Got SKIPPED message");
    return this.reject();
  }

  this.success();
}

func (cel *CustomEventListener) privateChannel(event: Event) {
  fmt.Println(`[consumer] Got a private message: "${event.body['message']}" ~> ${event.name}`);

  this.success();
}

fmt.Println("****************** Daemon Ready ******************");

Daemon.Start()
```

[See more details](https://github.com/pin-people/event_people_node/blob/master/examples/daemon.rb)

## Development

After checking out the repo, run `bin/setup` to install dependencies. Then, run `bin/test` to run the tests.

To install this module onto your local machine, run `go get`.

## Contributing

- Fork it
- Create your feature branch (`git checkout -b my-new-feature`)
- Commit your changes (`git commit -am 'Add some feature'`)
- Push to the branch (`git push origin my-new-feature`)
- Create a new Pull Request

## License

The module is available as open source under the terms of the [LGPL 3.0 License](https://www.gnu.org/licenses/lgpl-3.0.en.html).
