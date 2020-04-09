package events

import log "github.com/sirupsen/logrus"

// if subscribed to GroupIDAny
// any event will be received
const GroupIDAny GroupID = "*"

type EventType string
type GroupID string
type EventPayload interface{}

type event struct {
	Type    EventType
	GroupID GroupID
	Data    EventPayload
}

// subscription contains subscription info
// and message channel
type subscription struct {
	Types   []EventType
	Groups  []GroupID
	Channel chan event
}

// todo: add unsubscribe method
type EventBus interface {
	Start()
	Stop()
	Subscribe([]EventType, []GroupID) subscription
	Unsubscribe(subscription)
	Publish(EventType, GroupID, EventPayload)
}

// eventBus stores the information about subscribers
// listening to specific event types and group ids
type eventBus struct {
	subscribers      map[EventType]map[GroupID]map[chan event]bool
	newSubscriptions chan subscription
	eventChan        chan event
	quit             chan struct{}
}

var _ EventBus = (*eventBus)(nil)

func NewEventBus() EventBus {
	return &eventBus{
		subscribers:      make(map[EventType]map[GroupID]map[chan event]bool),
		newSubscriptions: make(chan subscription),
		eventChan:        make(chan event),
		quit:             make(chan struct{}),
	}
}

func (eb *eventBus) Start() {
	go eb.loop()
	log.Info("[eventbus] successfully started")
}

func (eb *eventBus) Stop() {
	close(eb.quit)
}

func (eb *eventBus) Subscribe(types []EventType, groupIDs []GroupID) subscription {
	subscription := subscription{
		Types:   types,
		Groups:  groupIDs,
		Channel: make(chan event),
	}
	eb.newSubscriptions <- subscription
	return subscription
}

func (eb *eventBus) Unsubscribe(sub subscription) {
	for _, eventType := range sub.Types {
		if groups, ok := eb.subscribers[eventType]; ok {
			for _, id := range sub.Groups {
				delete(groups[id], sub.Channel)

				if len(groups[id]) == 0 {
					delete(groups, id)
				}
			}

			if len(groups) == 0 {
				delete(eb.subscribers, eventType)
			}
		}
	}
}

func (eb *eventBus) Publish(eventType EventType, groupID GroupID, data EventPayload) {
	ev := event{
		Type:    eventType,
		GroupID: groupID,
		Data:    data,
	}
	eb.eventChan <- ev
}

func (eb *eventBus) loop() {
	for {
		select {

		case sub := <-eb.newSubscriptions:
			eb.subscribe(sub)

		case ev := <-eb.eventChan:
			eb.forwardEvent(ev)

		case <-eb.quit:
			log.Info("[eventbus] stopped")
			return

		}
	}
}

func (eb *eventBus) forwardEvent(ev event) {
	msg := "[eventbus] forward event"

	log.Infof("%v: received event: type={%v} groupID={%v}", msg, ev.Type, ev.GroupID)

	broadcastList := make(map[chan event]bool)

	if groups, typeExists := eb.subscribers[ev.Type]; typeExists {
		if chans, ok := groups[ev.GroupID]; ok {
			// add channels in this group to broadcast list
			for ch := range chans {
				broadcastList[ch] = true
			}
		}

		// send event to subscribers that listen to all groups
		if chans, ok := groups[GroupIDAny]; ok {
			for ch := range chans {
				broadcastList[ch] = true
			}
		}
	}

	// broadcast in goroutine to avoid blocking
	go func(channels map[chan event]bool, ev event) {
		for ch := range channels {
			ch <- ev
		}
		log.Infof("%v: type={%v} groupID={%v} to %v clients", msg, ev.Type, ev.GroupID, len(channels))
	}(broadcastList, ev)
}

func (eb *eventBus) subscribe(sub subscription) {
	for _, evType := range sub.Types {
		groups, ok := eb.subscribers[evType]
		if !ok {
			groups = make(map[GroupID]map[chan event]bool)
			eb.subscribers[evType] = groups
		}

		for _, id := range sub.Groups {
			chans, ok := groups[id]
			if !ok {
				chans = make(map[chan event]bool)
				groups[id] = chans
			}

			chans[sub.Channel] = true
		}
	}

	log.Infof("[eventbus] new subscription: types=%v groups=%v", sub.Types, sub.Groups)
}
