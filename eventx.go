package eventbusx

import "context"

type EventX[T any] struct {
	value T
	ctx   context.Context
}

func NewEventX[T any](v T) *EventX[T] {
	return &EventX[T]{
		value: v,
	}
}

func (e *EventX[T]) Value() T {
	return e.value
}

// Context returns the message's context. To change the context, use
// SetContext.
//
// The returned context is always non-nil; it defaults to the
// background context.
func (m *EventX[T]) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	return context.Background()
}

// SetContext sets provided context to the message.
func (m *EventX[T]) SetContext(ctx context.Context) {
	m.ctx = ctx
}

type AnyEvent = EventX[interface{}]

// Messages is a slice of messages.
type EventXList = []*AnyEvent
