// Sources for https://watermill.io/docs/getting-started/
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/abmpio/eventbusx"
	"github.com/abmpio/eventbusx/message"
	"github.com/abmpio/eventbusx/pubsub/gochannel"
)

func main() {
	subscribeEvent()
	// publishEvent()
	asyncPublishEvent()
}

func testMessages() {
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{},
		eventbusx.NewStdLogger(false, false),
	)

	messages, err := pubSub.Subscribe(context.Background(), "example.topic")
	if err != nil {
		panic(err)
	}

	go process(messages)

	publishMessages(pubSub)
}

func publishMessages(publisher message.Publisher) {
	for {
		msg := message.NewMessage(eventbusx.NewUUID(), []byte("Hello, world!"))

		if err := publisher.Publish("example.topic", msg); err != nil {
			panic(err)
		}

		time.Sleep(time.Second)
	}
}

func process(messages <-chan *message.Message) {
	for msg := range messages {
		fmt.Printf("received message: %s, payload: %s\n", msg.UUID, string(msg.Payload))

		// we need to Acknowledge that we received and processed the message,
		// otherwise, it will be resent over and over again.
		msg.Ack()
	}
}

type packet struct {
	M string
}

func publishEvent() {
	for {
		eventbusx.Publish(&packet{
			M: fmt.Sprintf("this is test event,uuid:%s", eventbusx.NewUUID()),
		})
		time.Sleep(time.Second)
	}
}

func asyncPublishEvent() {
	for {
		eventbusx.PublishAsync(&packet{
			M: fmt.Sprintf("this is test event,uuid:%s", eventbusx.NewUUID()),
		})
		time.Sleep(time.Second)
	}
}

func subscribeEvent() {
	eventbusx.SubscribeWithAction[packet](func(e *eventbusx.EventArgs) {
		fmt.Println("received events,observer1", eventbusx.EventArgsValueAs[packet](e))
		time.Sleep(time.Second * 10)
		panic("error")
	})
	eventbusx.SubscribeWithAction[packet](func(e *eventbusx.EventArgs) {
		fmt.Println("received events,observer2", eventbusx.EventArgsValueAs[packet](e))
		// time.Sleep(time.Second * 5)
		// panic("error")
	})
}
