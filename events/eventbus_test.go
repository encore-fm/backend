package events

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus()
	bus.Start()
	defer bus.Stop()

	sub := bus.Subscribe([]EventType{"event1", "event2"}, []GroupID{"group1"})
	bus.Publish("event1", "group1", "data")

	// block until message received
	<-sub.Channel
	_, ok := bus.(*eventBus).subscribers["event1"]["group1"][sub.Channel]
	assert.True(t, ok)
	_, ok = bus.(*eventBus).subscribers["event2"]["group1"][sub.Channel]
	assert.True(t, ok)

	bus.Unsubscribe(sub)
	for {
		if _, ok := <-sub.Channel; !ok {
			break
		}
		time.Sleep(time.Millisecond * 200)
	}

	_, ok = bus.(*eventBus).subscribers["event1"]
	assert.False(t, ok)
	_, ok = bus.(*eventBus).subscribers["event2"]
	assert.False(t, ok)
}

func TestEventBus_Publish(t *testing.T) {
	eventBus := NewEventBus()
	eventBus.Start()
	defer eventBus.Stop()

	eventSplice := []Event{
		{
			Type:    "event1",
			GroupID: "group1",
			Data:    "hello1",
		},
		{
			Type:    "event2",
			GroupID: "group1",
			Data:    "hello2",
		},
		{
			Type:    "event3",
			GroupID: "group1",
			Data:    "hello3",
		},
	}

	sub1 := eventBus.Subscribe([]EventType{"event1", "event2"}, []GroupID{"group1"})
	ch1 := make(chan []Event)

	go func() {
		var collector []Event

		for {
			res, ok := <-sub1.Channel
			if !ok {
				ch1 <- collector
				return
			}
			collector = append(collector, res)
		}

	}()

	for _, ev := range eventSplice {
		eventBus.Publish(ev.Type, ev.GroupID, ev.Data)
	}

	// sleep 1 second to allow every event to be forwarded
	time.Sleep(1 * time.Second)

	eventBus.Unsubscribe(sub1)
	events1 := <-ch1

	assert.Equal(t, 2, len(events1))
}
