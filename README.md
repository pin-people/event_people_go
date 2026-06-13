# EventPeople — Go

[![CircleCI](https://dl.circleci.com/status-badge/img/gh/pin-people/event_people_go/tree/main.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/pin-people/event_people_go/tree/main)

EventPeople is a tool to simplify the communication of event-based services. It is based on the [EventBus](https://github.com/EmpregoLigado/event_bus_rb) gem.

The main idea is to provide a tool that can emit or consume events based on their names. The event name has 4 words (`resource.origin.action.destination`) which defines important information about what kind of event it is, where it comes from, and who is eligible to consume it:

- **resource:** Defines which resource this event is related to — a `user`, a `product`, `company`, or anything you want;
- **origin:** Defines the name of the system which emitted the event;
- **action:** What action was performed on the resource — `created`, `deleted`, `updated`, etc.;
- **destination (Optional):** If not provided, EventPeople appends `.all`. It defines which service should consume the event. If set to a specific app name, only that service receives it (useful for replaying events).

As of today EventPeople uses RabbitMQ as its broker. Support for other brokers may be added in the future.

Spec version: **1.2.0**

## Installation

Add this line to your application's `go.mod`:

```yaml
require github.com/pin-people/event_people_go
```

To install and add it as a dependency:

```cmd
go get "github.com/pin-people/event_people_go"
```

### Linux/MacOS (proxy mode)

```cmd
GOPROXY=https://proxy.golang.org GO111MODULE=on go get github.com/pin-people/event_people_go
```

### Windows (proxy mode)

```cmd
set GOPROXY=https://proxy.golang.org; set GO111MODULE=on; go get github.com/pin-people/event_people_go
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `RABBIT_EVENT_PEOPLE_APP_NAME` | yes | Application name — used as queue name prefix |
| `RABBIT_EVENT_PEOPLE_TOPIC_NAME` | yes | RabbitMQ topic exchange name |
| `RABBIT_EVENT_PEOPLE_VHOST` | yes | RabbitMQ virtual host |
| `RABBIT_URL` | yes | RabbitMQ connection URL (e.g. `amqp://user:pass@host`) |

Retry behaviour is configured via `Config.Configure()` or per-listener `RetryConfig` — not via env vars.

## Setup

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

### Retry Configuration

Retry behaviour can be configured globally or per-listener. When not configured, hardcoded defaults apply.

**Global configuration (optional):**

```golang
EventPeople.Config.Configure(EventPeople.RetryConfig{
    MaxAttempts:   5,
    InitialDelay:  2000,      // milliseconds
    DelayStrategy: "exponential", // "exponential" (default) or "fixed"
    DLQName:       "myapp_dlq",   // defaults to "{appName}_dlq"
})
```

**Inspect the active configuration:**

```golang
rc := EventPeople.Config.GetRetryConfig()
// rc.MaxAttempts, rc.InitialDelay, rc.DelayStrategy, rc.DLQName
```

**Defaults:**

| Setting | Default |
|---------|---------|
| `MaxAttempts` | `3` |
| `InitialDelay` | `1000` ms |
| `DelayStrategy` | `"exponential"` |
| `DLQName` | `"{appName}_dlq"` |

**Delay strategies:**

- `exponential`: `min(initialDelay * 5^attempt, 600000 ms)`
  - attempt 0: 1 s, attempt 1: 5 s, attempt 2: 25 s, ...
- `fixed`: constant `initialDelay` on every retry

### RabbitMQ Topology

Declared automatically on subscribe:

| Resource | Name | Notes |
|----------|------|-------|
| Main queue | `{appName}-{resource}.{origin}.{action}.all` | `x-dead-letter-exchange` → DLX |
| Retry queue | `{queueName}_retry` | No queue TTL — per-message `expiration` |
| DLX exchange | `{appName}_dlx` | fanout, durable |
| DLQ | `{appName}_dlq` | bound to DLX |

### Events

The main component is `EventPeople.Event`. It wraps all event logic.

```golang
import EventPeople "github.com/pin-people/event_people_go/lib/event_people"

type Body struct {
    Amount int    `json:"amount"`
    Name   string `json:"name"`
}

func main() {
    event := EventPeople.NewEvent("user.users.created", Body{Amount: 42, Name: "John Doe"})
    // event.RetryCount == 0 (auto-initialized)
}
```

### Using the Emitter

```golang
import (
    EventPeople "github.com/pin-people/event_people_go/lib/event_people"
)

func main() {
    event := EventPeople.NewEvent("receipt.payments.pay.users", Body{Amount: 350, Name: "John"})
    EventPeople.TriggerEmitter([]*EventPeople.Event{event})
    EventPeople.Config.CloseConnection()
}
```

### Listening to Events

Three methods for processing results are available on the context:

- `Success()` — event processed successfully, ack and discard;
- `Fail()` — processing failed, publish to retry queue (with backoff) or nack to DLQ when retries exhausted;
- `Reject()` — discard without retrying, nack to DLQ.

The context also exposes:

- `ctx.GetMaxRetries()` — total retry attempts configured;
- `ctx.GetIsLastRetry()` — `true` on the final retry attempt.

**Single event listener:**

```golang
func main() {
    var eventName = "payment.payments.pay"
    var once = make(chan int)

    EventPeople.ListenTo(eventName, func(event EventPeople.Event, ctx EventPeople.ContextInterface) {
        fmt.Printf("Received %s: %s\n", event.Name, event.Body)

        if ctx.GetIsLastRetry() {
            fmt.Println("This is the last retry attempt")
        }
        ctx.Success()
        once <- 1
    })

    <-once
    EventPeople.Config.CloseConnection()
}
```

**Daemon with multiple event bindings:**

```golang
func pay(event EventPeople.Event, ctx EventPeople.ContextInterface) {
    fmt.Printf("Processing payment: %s\n", event.Name)
    ctx.Success()
}

func receive(event EventPeople.Event, ctx EventPeople.ContextInterface) {
    fmt.Printf("Received: %s\n", event.Name)
    ctx.Success()
}

func main() {
    EventPeople.BindEvent(pay, "resource.custom.pay")
    EventPeople.BindEvent(receive, "resource.custom.receive")

    EventPeople.DaemonStart()
}
```

**Per-listener retry configuration:**

Embed `BaseListener` in your listener struct and set `RetryConfig` before registering:

```golang
type PaymentListener struct {
    EventPeople.BaseListener
}

func main() {
    payListener := &PaymentListener{}
    payListener.RetryConfig = EventPeople.ListenerRetryConfig{
        MaxAttempts:   10,
        InitialDelay:  500,
        DelayStrategy: "fixed",
        DLQName:       "payments_dlq",
    }

    EventPeople.BindEvent(payListener.HandlePayment, "receipt.payments.pay")
    EventPeople.DaemonStart()
}
```

Per-listener settings override global `Config` defaults for that listener only. Zero values fall back to the global default.

### RetryCount on Incoming Events

When an event is redelivered from the retry queue, `event.RetryCount` is populated from the `x-event-people-retries` AMQP header:

```golang
EventPeople.ListenTo("payment.payments.pay", func(event EventPeople.Event, ctx EventPeople.ContextInterface) {
    fmt.Printf("Attempt %d of %d\n", event.RetryCount+1, ctx.GetMaxRetries())
    // process...
    ctx.Success()
})
```

## Retry and Dead Letter Queue (DLQ)

### Environment variables

-   `RABBIT_EVENT_PEOPLE_MAX_RETRIES` — max retry attempts before dead-lettering (default: `3`)
-   `RABBIT_EVENT_PEOPLE_RETRY_TTL_MS` — base delay in ms for retry backoff (default: `1000`)

### How it works

On `context.Fail()`:

-   If retries remain → message published to `{queue}_retry` with exponential backoff delay, then acked
-   If retries exhausted → nacked to DLQ via RabbitMQ DLX

On `context.Reject()` → nacked directly to DLQ (no retries)

**Delay strategies:**

-   `exponential` (default): `min(initialDelay × 5^retryCount, 600000)` ms
-   `fixed`: constant `initialDelay` ms

### Queue topology (auto-created on subscribe)

| Queue/Exchange | Name | Purpose |
|---|---|---|
| Exchange (DLX) | `{appName}_dlx` | Fanout, receives dead-lettered messages |
| DLQ | `{appName}_dlq` | Final resting place for failed messages |
| Retry queue | `{queue_name}_retry` | Holds messages until backoff delay expires |

### Usage

```go
import "github.com/pinpeople/event_people_go/lib/event_people"

func HandleOrder(event event_people.Event, context event_people.ContextInterface) {
    fmt.Printf("Attempt %d of %d\n", event.RetryCount+1, context.MaxRetries)

    if isInvalid(event) {
        context.Reject() // → DLQ immediately, no retries
        return
    }

    if err := process(event); err != nil {
        if context.IsLastRetry() {
            fmt.Println("Final attempt failed, sending to DLQ")
        }
        context.Fail() // → retry queue (or DLQ if exhausted)
        return
    }
    context.Success()
}

// Per-listener retry config
event_people.Listener.On("order.service.created", HandleOrder,
    event_people.RetryConfig{MaxAttempts: 5, DelayStrategy: "exponential"})
```

## Development

```cmd
go test ./...
```

## Contributing

- Fork it
- Create your feature branch (`git checkout -b my-new-feature`)
- Commit your changes (`git commit -am 'Add some feature'`)
- Push to the branch (`git push origin my-new-feature`)
- Create a new Pull Request

## License

The module is available as open source under the terms of the [LGPL 3.0 License](https://www.gnu.org/licenses/lgpl-3.0.en.html).
