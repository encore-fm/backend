package handlers

import "github.com/antonbaumann/spotify-jukebox/sse"

// Send event to broker
func (h *handler) SendEvent(sessionID string, eventType sse.EventType, data interface{}){
	event := sse.Event{
		GroupID: sessionID,
		Event:   eventType,
		Data:    data,
	}
	h.Broker.Notifier <- event
}