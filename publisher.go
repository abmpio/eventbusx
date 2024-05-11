package eventbusx

import (
	"context"
	"errors"
	"sync"

	"github.com/lithammer/shortuuid/v4"
)

type publisher struct {
	logger                 LoggerAdapter
	subscribersWg          sync.WaitGroup
	subscribers            map[string][]*subscriber
	subscribersLock        sync.RWMutex
	subscribersByTopicLock sync.Map // map of *sync.Mutex

	closed     bool
	closedLock sync.Mutex
	closing    chan struct{}
}

func newPublisher() *publisher {
	return &publisher{
		logger: _globalLogger.With(LogFields{
			"pubsub_uuid": shortuuid.New(),
		}),
		subscribers:            make(map[string][]*subscriber),
		subscribersByTopicLock: sync.Map{},

		closing: make(chan struct{}),
	}
}

func (g *publisher) Publish(topic string, events ...*AnyEvent) error {
	if g.isClosed() {
		return errors.New("Pub/Sub closed")
	}

	g.subscribersLock.RLock()
	defer g.subscribersLock.RUnlock()

	subLock, _ := g.subscribersByTopicLock.LoadOrStore(topic, &sync.Mutex{})
	subLock.(*sync.Mutex).Lock()
	defer subLock.(*sync.Mutex).Unlock()

	for i := range events {
		currentEventX := events[i]

		_, err := g.sendMessage(topic, currentEventX)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *publisher) sendMessage(topic string, message *AnyEvent) (<-chan struct{}, error) {
	subscribers := g.topicSubscribers(topic)
	ackedBySubscribers := make(chan struct{})

	logFields := LogFields{"topic": topic}

	if len(subscribers) == 0 {
		close(ackedBySubscribers)
		g.logger.Info("No subscribers to send message", logFields)
		return ackedBySubscribers, nil
	}

	go func(subscribers []*subscriber) {
		wg := &sync.WaitGroup{}

		for i := range subscribers {
			subscriber := subscribers[i]

			wg.Add(1)
			go func() {
				subscriber.sendMessageToSubscriber(message, logFields)
				wg.Done()
			}()
		}

		wg.Wait()
		close(ackedBySubscribers)
	}(subscribers)

	return ackedBySubscribers, nil
}

func (g *publisher) Subscribe(ctx context.Context, topic string) (<-chan *AnyEvent, error) {
	g.closedLock.Lock()

	if g.closed {
		g.closedLock.Unlock()
		return nil, errors.New("Pub/Sub closed")
	}

	g.subscribersWg.Add(1)
	g.closedLock.Unlock()

	g.subscribersLock.Lock()

	subLock, _ := g.subscribersByTopicLock.LoadOrStore(topic, &sync.Mutex{})
	subLock.(*sync.Mutex).Lock()

	s := &subscriber{
		ctx:           ctx,
		outputChannel: make(chan *AnyEvent),
		logger:        g.logger,
		closing:       make(chan struct{}),
	}

	go func(s *subscriber, g *publisher) {
		select {
		case <-ctx.Done():
			// unblock
		case <-g.closing:
			// unblock
		}

		s.Close()

		g.subscribersLock.Lock()
		defer g.subscribersLock.Unlock()

		subLock, _ := g.subscribersByTopicLock.Load(topic)
		subLock.(*sync.Mutex).Lock()
		defer subLock.(*sync.Mutex).Unlock()

		g.removeSubscriber(topic, s)
		g.subscribersWg.Done()
	}(s, g)

	defer g.subscribersLock.Unlock()
	defer subLock.(*sync.Mutex).Unlock()

	g.addSubscriber(topic, s)

	return s.outputChannel, nil
}

func (g *publisher) addSubscriber(topic string, s *subscriber) {
	if _, ok := g.subscribers[topic]; !ok {
		g.subscribers[topic] = make([]*subscriber, 0)
	}
	g.subscribers[topic] = append(g.subscribers[topic], s)
}

func (g *publisher) removeSubscriber(topic string, toRemove *subscriber) {
	removed := false
	for i, sub := range g.subscribers[topic] {
		if sub == toRemove {
			g.subscribers[topic] = append(g.subscribers[topic][:i], g.subscribers[topic][i+1:]...)
			removed = true
			break
		}
	}
	if !removed {
		panic("cannot remove subscriber, not found ")
	}
}

func (g *publisher) topicSubscribers(topic string) []*subscriber {
	subscribers, ok := g.subscribers[topic]
	if !ok {
		return nil
	}

	// let's do a copy to avoid race conditions and deadlocks due to lock
	subscribersCopy := make([]*subscriber, len(subscribers))
	copy(subscribersCopy, subscribers)

	return subscribersCopy
}

func (p *publisher) isClosed() bool {
	p.closedLock.Lock()
	defer p.closedLock.Unlock()

	return p.closed
}

// Close closes the GoChannel Pub/Sub.
func (g *publisher) Close() error {
	g.closedLock.Lock()
	defer g.closedLock.Unlock()

	if g.closed {
		return nil
	}

	g.closed = true
	close(g.closing)

	g.logger.Debug("Closing Pub/Sub, waiting for subscribers", nil)
	g.subscribersWg.Wait()
	g.logger.Info("Pub/Sub closed", nil)

	return nil
}
