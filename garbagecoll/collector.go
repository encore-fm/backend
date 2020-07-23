package garbagecoll

import (
	"context"
	"github.com/antonbaumann/spotify-jukebox/db"
	"github.com/sirupsen/logrus"
	"time"
)

// responsible for deleting inactive sessionCollection
type garbageCollector struct {
	sessionExpiration time.Duration
	ticker            *time.Ticker
	userCollection    db.UserCollection
	sessionCollection db.SessionCollection
	quit              chan bool
}

func New(
	sessionExpiration time.Duration,
	cleanInterval time.Duration,
	users db.UserCollection,
	sessions db.SessionCollection,
) *garbageCollector {
	return &garbageCollector{
		sessionExpiration: sessionExpiration,
		ticker:            time.NewTicker(cleanInterval),
		userCollection:    users,
		sessionCollection: sessions,
	}
}

func (gc garbageCollector) Run() {
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

func (gc garbageCollector) Quit() {
	if gc.quit == nil {
		logrus.Warn("garbage collector not running")
		return
	}
	gc.quit <- true
}

func (gc garbageCollector) clean() {
	msg := "[garbagecoll] clean"
	ctx := context.Background()

	sessionIDs, err := gc.sessionCollection.ListSessionIDs(ctx)
	if err != nil {
		logrus.Warnf("%v, %v", msg, err)
		return
	}

	for _, sessionID := range sessionIDs {
		session, err := gc.sessionCollection.GetSessionByID(ctx, sessionID)
		if err != nil {
			logrus.Warnf("%v, %v", msg, err)
			continue
		}
		if time.Since(session.LastUpdated) > gc.sessionExpiration {
			err = gc.userCollection.DeleteUsersBySessionID(ctx, sessionID)
			if err != nil {
				logrus.Warnf("%v, %v", msg, err)
				continue
			}
			err = gc.sessionCollection.DeleteSession(ctx, sessionID)
			if err != nil {
				logrus.Warnf("%v, %v", msg, err)
				continue
			}
		}
	}
}
