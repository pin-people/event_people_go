package EventPeople

import (
	"os"
	"strings"
)

// ListenerRetryConfig holds per-listener retry overrides (class-level attributes).
// Zero values mean "use Config defaults".
type ListenerRetryConfig struct {
	MaxAttempts   int
	InitialDelay  int
	DelayStrategy string
	DLQName       string
}

// AbstractBaseListener defines the minimal interface all listeners must satisfy.
type AbstractBaseListener interface {
	Initialize(context ContextInterface)
	Success()
	Fail()
	Reject()
}

// BaseListener is the base class for user-defined listener classes.
// Embed this struct in your listener to inherit Success/Fail/Reject delegation.
//
// Class-level retry configuration can be set via the RetryConfig field before
// registering the listener. All fields are optional and fall back to Config defaults.
type BaseListener struct {
	AbstractBaseListener
	// context is the ContextInterface injected by Queue.callback before dispatch.
	context ContextInterface
	// RetryConfig holds per-listener retry overrides (class-level static attributes).
	RetryConfig ListenerRetryConfig
}

// Initialize sets the context on the listener instance.
func (base *BaseListener) Initialize(context ContextInterface) {
	base.setContext(context)
}

func (base *BaseListener) setContext(context ContextInterface) {
	base.context = context
}

// ResolvedRetryConfig returns the effective retry configuration for this listener,
// merging listener-level overrides with global Config defaults.
func (base *BaseListener) ResolvedRetryConfig() RetryConfig {
	global := Config.GetRetryConfig()

	maxAttempts := global.MaxAttempts
	if base.RetryConfig.MaxAttempts != 0 {
		maxAttempts = base.RetryConfig.MaxAttempts
	}

	initialDelay := global.InitialDelay
	if base.RetryConfig.InitialDelay != 0 {
		initialDelay = base.RetryConfig.InitialDelay
	}

	delayStrategy := global.DelayStrategy
	if base.RetryConfig.DelayStrategy != "" {
		delayStrategy = base.RetryConfig.DelayStrategy
	}

	dlqName := global.DLQName
	if base.RetryConfig.DLQName != "" {
		dlqName = base.RetryConfig.DLQName
	}

	return RetryConfig{
		MaxAttempts:   maxAttempts,
		InitialDelay:  initialDelay,
		DelayStrategy: delayStrategy,
		DLQName:       dlqName,
	}
}

// Success delegates to context.Success().
func (base *BaseListener) Success() {
	base.context.Success()
}

// Fail delegates to context.Fail().
func (base *BaseListener) Fail() {
	base.context.Fail()
}

// Reject delegates to context.Reject().
func (base *BaseListener) Reject() {
	base.context.Reject()
}

// BindEvent registers a listener method for an event name.
//
// When eventName has 3 parts, the lib creates ONE queue
// ({appName}-resource.origin.action.all) with TWO exchange bindings:
//   - routing key resource.origin.action.all  (broadcast)
//   - routing key resource.origin.action.{appName}  (targeted)
//
// A second physical queue is NOT created — queueNameByRoutingKey normalizes the
// destination to "all" in both cases (DEV-GO-002 fix).
func BindEvent(method ListenerMethod, eventName string) {
	eventNameSplited := strings.Split(eventName, ".")

	if len(eventNameSplited) <= 3 {
		ListenerManager.Register(
			ListenerManagerStruct{
				EventName: FixedEventName(eventName, "all"),
				Method:    method,
				Listener:  &BaseListener{},
			},
		)
		ListenerManager.Register(
			ListenerManagerStruct{
				EventName: FixedEventName(eventName, os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME")),
				Method:    method,
				Listener:  &BaseListener{},
			},
		)
		return
	}
	ListenerManager.Register(
		ListenerManagerStruct{
			EventName: FixedEventName(eventName, os.Getenv("RABBIT_EVENT_PEOPLE_APP_NAME")),
			Method:    method,
			Listener:  &BaseListener{},
		},
	)
}

// FixedEventName returns the event name with postfix applied to the destination segment.
func FixedEventName(eventName string, postfix string) string {
	eventNameSplited := strings.Split(eventName, ".")

	if len(eventNameSplited) == 4 {
		eventNameSplited[3] = postfix
		return strings.Join(eventNameSplited, ".")
	}
	eventNameSplited = append(eventNameSplited, postfix)
	return strings.Join(eventNameSplited, ".")
}
