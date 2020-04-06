package player

import (
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

type EventType string

const (
	AdminStateChangedEvent EventType = "admin_player_state_changed"
	UserStateChangedEvent EventType = "user_player_state_changed"
)

type StateChangedPayload struct {
	SongID   string `json:"current_track"`
	Duration int    `json:"duration"`
	Position int    `json:"position"`
	Paused   bool   `json:"paused"`
}

type Event struct {
	SessionID string
	Type      EventType
	Payload   interface{}
}

type Controller struct {
	Events chan Event
}

func NewController() *Controller {
	controller := &Controller{Events: make(chan Event)}
	return controller
}

func (ctrl *Controller) Start() {
	go ctrl.eventLoop()
}

func (ctrl *Controller) eventLoop() {
	for {
		select {
		case event := <-ctrl.Events:
			switch event.Type {
			case AdminStateChangedEvent:
				log.Info("[playerctrl] AdminStateChangeEvent")
				spew.Dump(event)
			case UserStateChangedEvent:
				log.Info("[playerctrl] UserStateChangeEvent")
				spew.Dump(event)
			default:
				spew.Dump(event)
			}
		}
	}
}
