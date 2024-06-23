package eventbusx

import (
	"unsafe"

	"github.com/abmpio/x/factory"
)

var (
	_globalPublisher *publisher = newPublisher()
)

// publish events, synchronized
func Publish[T any](v *T) {
	if v == nil {
		return
	}
	topic := factory.ParseUnderlyTypeId(new(T))
	var arg *interface{} = (*interface{})(unsafe.Pointer(v))
	_globalPublisher.notifyObserver(topic, NewEventArgs(arg), false)
}

// publish events with result, synchronized
func PublishWithResult[T any](v *T) error {
	if v == nil {
		return nil
	}
	topic := factory.ParseUnderlyTypeId(new(T))
	var arg *interface{} = (*interface{})(unsafe.Pointer(v))
	return _globalPublisher.notifyObserverWithResult(topic, NewEventArgs(arg), false)
}

// publish events, asynchronized
func PublishAsync[T any](v *T) {
	if v == nil {
		return
	}
	topic := factory.ParseUnderlyTypeId(new(T))
	var arg *interface{} = (*interface{})(unsafe.Pointer(v))
	_globalPublisher.notifyObserver(topic, NewEventArgs(arg), true)
}

func Subscribe[T any](observer IEventObserver) {
	topic := factory.ParseUnderlyTypeId(new(T))
	_globalPublisher.RegistObserver(topic, observer)
}

func SubscribeWithAction[T any](fn func(v *EventArgs) error) {
	topic := factory.ParseUnderlyTypeId(new(T))
	observer := EventObserverFromAction(fn)
	_globalPublisher.RegistObserver(topic, observer)
}
