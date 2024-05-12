package eventbusx

import (
	"context"
	"unsafe"
)

type EventArgs struct {
	value *interface{}
	ctx   context.Context
}

func NewEventArgs(v *interface{}) *EventArgs {
	return &EventArgs{
		value: v,
	}
}

func EventArgsValueAs[T any](e *EventArgs) *T {
	if e == nil {
		return nil
	}
	var v *T = (*T)(unsafe.Pointer(e.value))
	return v
}

// Context returns the message's context. To change the context, use
// SetContext.
//
// The returned context is always non-nil; it defaults to the
// background context.
func (m *EventArgs) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	return context.Background()
}

// SetContext sets provided context to the message.
func (m *EventArgs) SetContext(ctx context.Context) {
	m.ctx = ctx
}
