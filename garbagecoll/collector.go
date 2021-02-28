package garbagecoll

import (
	"context"
	"time"

	"github.com/encore-fm/backend/config"
	"github.com/encore-fm/backend/db"
	"github.com/encore-fm/backend/events"
	"github.com/sirupsen/logrus"
)

type GarbageCollector interface {
	Start()
	Stop()
}

// responsible for deleting inactive sessions
type garbageCollector struct {
	ticker            *time.Ticker
	sessionExpiration time.Duration
	userCollection    db.UserCollection
	sessionCollection db.SessionCollection
	eventBus          events.EventBus
	quit              chan bool
}

var _ GarbageCollector = (*garbageCollector)(nil)

func New(
	users db.UserCollection,
	sessions db.SessionCollection,
	eventBus events.EventBus,
) GarbageCollector {
	cleaningInterval := time.Second * time.Duration(config.Conf.GarbageCollector.CleaningIntervalInS)
	sessionExpiration := time.Second * time.Duration(config.Conf.GarbageCollector.SessionExpirationInS)
	return &garbageCollector{
		ticker:            time.NewTicker(cleaningInterval),
		sessionExpiration: sessionExpiration,
		userCollection:    users,
		sessionCollection: sessions,
		eventBus:          eventBus,
	}
}

func (gc garbageCollector) Start() {
	// quit chanel should only exist when gc is running
	quit := make(chan bool)
	gc.quit = quit

	go func() {
		for {
			select {
			case <-gc.ticker.C:
				gc.clean()
			case <-quit:
				gc.ticker.Stop()
				gc.quit = nil
				return
			}
		}
	}()
}

func (gc garbageCollector) Stop() {
	if gc.quit == nil {
		logrus.Warn("garbage collector not running")
		return
	}
	gc.quit <- true
}

func (gc garbageCollector) clean() {
	msg := "[garbagecoll] clean"
	ctx := context.Background()

	expiredSessions, err := gc.sessionCollection.ListExpiredSessions(ctx, gc.sessionExpiration)
	if err != nil {
		logrus.Warnf("%v, %v", msg, err)
		return
	}
	if len(expiredSessions) == 0 {
		return
	}

	err = gc.userCollection.DeleteUsersBySessionIDs(ctx, expiredSessions)
	if err != nil {
		logrus.Warnf("%v, %v", msg, err)
		return
	}
	err = gc.sessionCollection.DeleteSessions(ctx, expiredSessions)
	if err != nil {
		logrus.Warnf("%v, %v", msg, err)
		return
	}

	gc.eventBus.RemoveGroups(events.AsGroupIDs(expiredSessions))

	logrus.Infof("deleted %v session(s)", len(expiredSessions))
}
