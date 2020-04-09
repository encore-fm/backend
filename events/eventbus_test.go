package events

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

// todo write better tests
func TestEventBus_Publish(t *testing.T) {
	eventBus := NewEventBus()
	eventBus.Start()
	defer eventBus.Stop()

	eventSplice := []event{
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
	ch1 := make(chan []event)

	go func() {
		var collector []event

		for i := 0; i < 2; i++ {
			res := <-sub1.Channel
			collector = append(collector, res)
		}

		ch1 <- collector
	}()

	for _, ev := range eventSplice {
		eventBus.Publish(ev.Type, ev.GroupID, ev.Data)
	}

	events1 := <-ch1

	assert.Equal(t, 2, len(events1))
}
