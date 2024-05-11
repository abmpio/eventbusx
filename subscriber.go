package eventbusx

import (
	"context"
	"sync"
)

type subscriber struct {
	ctx context.Context

	sending       sync.Mutex
	outputChannel chan *AnyEvent

	logger  LoggerAdapter
	closed  bool
	closing chan struct{}
}

func (s *subscriber) Close() {
	if s.closed {
		return
	}
	close(s.closing)

	s.logger.Debug("Closing subscriber, waiting for sending lock", nil)
	// ensuring that we are not sending to closed channel
	s.sending.Lock()
	defer s.sending.Unlock()

	s.logger.Debug("publisher Pub/Sub Subscriber closed", nil)
	s.closed = true

	close(s.outputChannel)
}

func (s *subscriber) sendMessageToSubscriber(msg *AnyEvent, logFields LogFields) {
	s.sending.Lock()
	defer s.sending.Unlock()

	ctx, cancelCtx := context.WithCancel(s.ctx)
	defer cancelCtx()

	for {
		// copy the message to prevent ack/nack propagation to other consumers
		// also allows to make retries on a fresh copy of the original message
		msgToSend := msg
		msgToSend.SetContext(ctx)

		s.logger.Trace("Sending msg to subscriber", logFields)
		if s.closed {
			s.logger.Info("Pub/Sub closed, discarding msg", logFields)
			return
		}

		select {
		case s.outputChannel <- msgToSend:
			s.logger.Trace("Sent message to subscriber", logFields)
		case <-s.closing:
			s.logger.Trace("Closing, message discarded", logFields)
			return
		}

		select {
		case <-s.closing:
			s.logger.Trace("Closing, message discarded", logFields)
			return
		}
	}
}
