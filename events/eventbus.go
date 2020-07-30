package events

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// if subscribed to GroupIDAny
// any Event will be received
const GroupIDAny GroupID = "*"

type EventType string
type GroupID string
type EventPayload interface{}

type Event struct {
	Type    EventType
	GroupID GroupID
	Data    EventPayload
}

// subscription contains subscription info
// and message channel
type subscription struct {
	Types   []EventType
	Groups  []GroupID
	Channel chan Event
}

type EventBus interface {
	Start()
	Stop()
	Subscribe([]EventType, []GroupID) subscription
	Unsubscribe(subscription)
	RemoveGroups([]GroupID)
	Publish(EventType, GroupID, EventPayload)
}

// eventBus stores the information about subscribers
// listening to specific Event types and group ids
type eventBus struct {
	subscribers      map[EventType]map[GroupID]map[chan Event]bool
	newSubscriptions chan subscription
	unsubscriptions  chan subscription
	cleanups         chan []GroupID
	eventChan        chan Event
	quit             chan struct{}
	mapMutex         sync.RWMutex
}

var _ EventBus = (*eventBus)(nil)

func NewEventBus() EventBus {
	return &eventBus{
		subscribers:      make(map[EventType]map[GroupID]map[chan Event]bool),
		newSubscriptions: make(chan subscription),
		unsubscriptions:  make(chan subscription),
		cleanups:         make(chan []GroupID),
		eventChan:        make(chan Event, 20),
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
		Channel: make(chan Event, 5),
	}
	eb.newSubscriptions <- subscription
	return subscription
}

// removes channel from topics
func (eb *eventBus) Unsubscribe(sub subscription) {
	eb.unsubscriptions <- sub
}

// removes channel from outdated groups
func (eb *eventBus) RemoveGroups(groups []GroupID) {
	eb.cleanups <- groups
}

func (eb *eventBus) Publish(eventType EventType, groupID GroupID, data EventPayload) {
	ev := Event{
		Type:    eventType,
		GroupID: groupID,
		Data:    data,
	}
	eb.eventChan <- ev
}

func (eb *eventBus) loop() {
	for {
		select {

		case unsub := <-eb.unsubscriptions:
			eb.unsubscribe(unsub)

		case groups := <-eb.cleanups:
			eb.removeGroups(groups)

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

func (eb *eventBus) forwardEvent(ev Event) {
	eb.mapMutex.RLock()

	msg := "[eventbus] forward Event"
	log.Infof("%v: received Event: type={%v} groupID={%v}", msg, ev.Type, ev.GroupID)

	broadcastList := make(map[chan Event]bool)

	if groups, typeExists := eb.subscribers[ev.Type]; typeExists {
		if chans, ok := groups[ev.GroupID]; ok {
			// add channels in this group to broadcast list
			for ch := range chans {
				broadcastList[ch] = true
			}
		}

		// send Event to subscribers that listen to all groups
		if chans, ok := groups[GroupIDAny]; ok {
			for ch := range chans {
				broadcastList[ch] = true
			}
		}
	}

	//broadcast in goroutine to avoid blocking
	go func(channels map[chan Event]bool, ev Event) {
		defer eb.mapMutex.RUnlock()
		for ch := range channels {
			select {
			case ch <- ev:
			case <-time.After(time.Millisecond * 300):
				log.Warnf("%v: channel is blocking -> skipping", msg)
			}
		}
		log.Infof("%v: type={%v} groupID={%v} to %v clients", msg, ev.Type, ev.GroupID, len(channels))
	}(broadcastList, ev)
}

func (eb *eventBus) subscribe(sub subscription) {
	eb.mapMutex.Lock()
	defer eb.mapMutex.Unlock()

	for _, evType := range sub.Types {
		groups, ok := eb.subscribers[evType]
		if !ok {
			groups = make(map[GroupID]map[chan Event]bool)
			eb.subscribers[evType] = groups
		}

		for _, id := range sub.Groups {
			chans, ok := groups[id]
			if !ok {
				chans = make(map[chan Event]bool)
				groups[id] = chans
			}

			chans[sub.Channel] = true
		}
	}

	log.Infof("[eventbus] new subscription: types=%v groups=%v", sub.Types, sub.Groups)
}

func (eb *eventBus) unsubscribe(sub subscription) {
	eb.mapMutex.Lock()
	defer eb.mapMutex.Unlock()

	msg := "[eventbus] unsubscribe"

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

	close(sub.Channel)
	log.Infof("%v: type=%v groups=%v", msg, sub.Types, sub.Groups)
}

func (eb *eventBus) removeGroups(groups []GroupID) {
	eb.mapMutex.Lock()
	defer eb.mapMutex.Unlock()

	msg := "[eventbus] removeGroups"

	for eventType, groupToChan := range eb.subscribers {
		for _, id := range groups {
			delete(groupToChan, id)
		}

		// if this eventType does not have subscribers anymore
		// delete key
		if len(eb.subscribers[eventType]) == 0 {
			delete(eb.subscribers, eventType)
		}
	}

	log.Infof("%v: groups=%v", msg, groups)
}
