package eventbusx

import "context"

var (
	_globalPublisher *publisher = newPublisher()
)

func Publish(topic string, events ...*AnyEvent) error {
	return _globalPublisher.Publish(topic, events...)
}

func Subscribe(ctx context.Context, topic string) (<-chan *AnyEvent, error) {
	return _globalPublisher.Subscribe(ctx, topic)
}
