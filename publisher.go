package eventbusx

import (
	"sync"

	"github.com/lithammer/shortuuid/v4"
)

type publisher struct {
	logger       LoggerAdapter
	rwLock       sync.RWMutex
	observerList map[string]*eventObserverList
}

func newPublisher() *publisher {
	return &publisher{
		logger: _globalLogger.With(LogFields{
			"pubsub_uuid": shortuuid.New(),
		}),
		rwLock:       sync.RWMutex{},
		observerList: map[string]*eventObserverList{},
	}
}

func (p *publisher) RegistObserver(topic string, observer IEventObserver) {
	if observer == nil {
		return
	}
	list := p.ensureObserverList(topic)
	list.registObserver(observer)
}

func (p *publisher) UnregistObserver(topic string, observer IEventObserver) {
	if observer == nil {
		return
	}
	list := p.getObserverList(topic)
	if list == nil {
		return
	}
	list.unregistObserver(observer)
}

func (p *publisher) notifyObserver(topic string, e *EventArgs, async bool) {
	observerList := p.getObserverList(topic)
	if observerList == nil {
		return
	}
	observerList.notifyObserver(e, async)
}

func (p *publisher) notifyObserverWithResult(topic string, e *EventArgs, async bool) error {
	observerList := p.getObserverList(topic)
	if observerList == nil {
		return nil
	}
	return observerList.notifyObserverWithResult(e, async)
}

func (p *publisher) getObserverList(topic string) *eventObserverList {
	p.rwLock.RLock()
	defer p.rwLock.RUnlock()
	list := p.observerList[topic]
	return list
}

func (p *publisher) ensureObserverList(topic string) *eventObserverList {
	p.rwLock.Lock()
	defer p.rwLock.Unlock()
	list, ok := p.observerList[topic]
	if !ok {
		list = newEventObserverList()
		p.observerList[topic] = list
	}
	return list
}
