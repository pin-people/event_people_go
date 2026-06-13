# EventPeople — Go Implementation

Spec version target: see `.event_people.yml` → `spec_version`
Spec contract: `spec/contract.yml` in the main `event_people` repository
GitHub: https://github.com/pin-people/event_people_go

## Project Structure

```
lib/event_people/
├── config.go
├── event.go
├── listener.go
├── emitter.go
├── daemon.go
├── base_broker.go     ← BaseBroker interface
├── rabbit_broker.go   ← RabbitBroker
├── context.go         ← ContextInterface
├── rabbit_context.go  ← RabbitContext
├── rabbit_topic.go    ← Topic
├── rabbit_queue.go    ← Queue
├── base_listener.go   ← BaseListener
├── listener_manager.go ← ListenersManager
├── callback.go
└── delivery_info.go
```

## Critical Rule — Spec Conformance

**Before implementing any new feature or changing existing behavior:**

1. Check `spec/contract.yml` in the main repo for the expected interface definition
2. Check `.event_people.yml` for known deviations already accepted in this implementation

If the change **aligns with the spec**: implement and update `.event_people.yml` status accordingly.

If the change **would deviate from the spec** (different method name, different signature,
different attribute, different behavior):

→ **STOP and ask the user:**
> "This change deviates from spec/contract.yml. Should we:
> 1. Update the spec first (via /update-spec in the main repo), then implement here?
> 2. Conform to the current spec instead?"

Never implement a deviation silently.

## Critical Bugs — Fix These First

These bugs cause events to never be delivered or routed incorrectly:

| ID | File | Description |
|----|------|-------------|
| BUG-GO-001 | `event.go:fixName()` | AppName prefix added to `event.Name` → routing key mismatch, events never delivered |
| BUG-GO-002 | `rabbit_queue.go:callback` | `eventMessage.Name = eventMessage.Headers.AppName` → wrong event name in every callback |
| BUG-GO-003 | `rabbit_queue.go:createQueueAndBind` | `eventNameSplited[3]` with no bounds check → panic on 3-part event names |

Full details with fix instructions in `.event_people.yml` → `bugs`.

## Known Deviations

- **DEV-GO-001**: `Daemon.bindSignals()` not implemented — no graceful shutdown on SIGTERM/SIGINT

## Routing Convention (Critical)

Per `spec/contract.yml → routing_convention`:
- Queue name: `{appName}-{resource}.{origin}.{action}.all` ← appName prefix here only
- Routing key: `{resource}.{origin}.{action}.{destination}` ← NO appName prefix

BUG-GO-001 violates this by adding the appName to the routing key.

## Pending Features (blocked on spec)

- `Event.retryCount` and `incrementRetryCount()`
- `Config.maxAttempts`, `delayStrategy`, `dlqName`, `getRetryConfig()`
- `Listener.on` retry params
- `Context.maxRetries`, `isLastRetry`
- `Daemon.bindSignals()`
- `RetryManager` component

Do not implement these until the spec marks them as `status: stable`.
