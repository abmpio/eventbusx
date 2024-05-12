package eventbusx

import (
	"fmt"
	"sync"
)

type IEventObserver interface {
	Notify(e *EventArgs)
}

type actionEventObserver struct {
	action func(e *EventArgs)
}

var _ IEventObserver = (*actionEventObserver)(nil)

func (a *actionEventObserver) Notify(e *EventArgs) {
	if a.action == nil {
		return
	}
	a.action(e)
}

func EventObserverFromAction(action func(*EventArgs)) IEventObserver {
	return &actionEventObserver{
		action: action,
	}
}

type eventObserverList struct {
	rwLock sync.RWMutex
	list   []IEventObserver
}

func newEventObserverList() *eventObserverList {
	return &eventObserverList{
		rwLock: sync.RWMutex{},
		list:   make([]IEventObserver, 0),
	}
}

func (l *eventObserverList) notifyObserver(e *EventArgs, async bool) {
	l.rwLock.RLock()
	cList := l.list
	l.rwLock.RUnlock()

	notifyFn := func() {
		for _, eachObserver := range cList {
			currentObserver := eachObserver
			notifyFn := func() {
				defer func() {
					if p := recover(); p != nil {
						msg := fmt.Sprint(p)
						_globalLogger.Info(fmt.Sprintf("panic when actionEventObserver notify, err:%s", msg), nil)
					}
				}()
				currentObserver.Notify(e)
			}
			notifyFn()
		}
	}
	if !async {
		notifyFn()
	} else {
		// async, go func
		go notifyFn()
	}
}

func (l *eventObserverList) registObserver(observers ...IEventObserver) {
	if len(observers) <= 0 {
		return
	}
	l.rwLock.Lock()
	l.list = append(l.list, observers...)
	l.rwLock.Unlock()
}

func (l *eventObserverList) unregistObserver(observer IEventObserver) {
	if observer == nil {
		return
	}
	l.rwLock.Lock()
	defer l.rwLock.Unlock()
	index := -1
	for i := range l.list {
		if l.list[i] == observer {
			index = i
			break
		}
	}
	l.list = l.removeByIndex(l.list, index)
}

func (l *eventObserverList) removeByIndex(v []IEventObserver, index int) []IEventObserver {
	return append(v[:index], v[index+1:]...)
}
