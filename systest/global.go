package systest

import (
	"fmt"
	"time"

	"github.com/antonbaumann/spotify-jukebox/player"
	"golang.org/x/oauth2"

	"github.com/antonbaumann/spotify-jukebox/session"
	"github.com/antonbaumann/spotify-jukebox/song"
	"github.com/antonbaumann/spotify-jukebox/user"
	"go.mongodb.org/mongo-driver/mongo"
)

const BackendBaseUrl = "http://127.0.0.1:8080"

var (
	client *mongo.Client

	sessionCollection *mongo.Collection
	userCollection    *mongo.Collection
)

const (
	TestSessionID = "91cb20eeb38943efb0ee22147f3a1b70d59d4d8e3b3c54771bb89ea8cce3f0c51b8e6ef4faab22e7f1c9e3db9c8c220eb26f5ed8936a63b63727648ea9963698"

	TestAdminUsername = "baumanto"
	TestAdminSecret   = "1234"

	TestUserName   = "omar"
	TestUserSecret = "1234"

	NotRickRollSongID = "4uLU6hMCjMI75M1A2tKUQC"
	SkiFoanID         = "3gnB6G7MCqB0xjYiAdiaSY"
	AntonAusTirolID   = "2YuKyP77pidQlkxm8PuyJj"
	CordulaSongID     = "483ykNWhSQXYprueXUMNeo"
)

var (
	testNow     = time.Now()
	testSession = &session.Session{
		ID: TestSessionID,
		SongList: []*song.Model{{
			ID:          SkiFoanID,
			SuggestedBy: TestAdminUsername,
			Score:       1,
			Upvoters:    []string{TestAdminUsername},
			Downvoters:  make([]string, 0),
			Duration:    10000,
		}, {
			ID:          AntonAusTirolID,
			SuggestedBy: TestUserName,
			Score:       1,
			Upvoters:    []string{TestAdminUsername, TestUserName},
			Downvoters:  []string{},
			Duration:    10000,
		}, {
			ID:          CordulaSongID,
			SuggestedBy: TestAdminUsername,
			Score:       1,
			Upvoters:    []string{TestAdminUsername},
			Downvoters:  []string{TestUserName},
			Duration:    10000,
		}},
		Player: &player.Player{
			CurrentSong: &song.Model{
				Name:     "Not Rock Roll Song",
				ID:       NotRickRollSongID,
				Duration: 10000, // 10 seconds
			},
			SongStart:     testNow.Add(-2 * time.Minute),
			PauseStart:    testNow.Add(-1 * time.Minute),
			PauseDuration: 30 * time.Second,
			Paused:        false,
		},
	}
	testAdmin = &user.Model{
		ID:                fmt.Sprintf("%v@%v", TestAdminUsername, TestSessionID),
		Username:          TestAdminUsername,
		Secret:            TestAdminSecret,
		SessionID:         TestSessionID,
		IsAdmin:           true,
		Score:             1,
		SpotifyAuthorized: true,
		AuthToken: &oauth2.Token{
			AccessToken:  "1234",
			TokenType:    "Bearer",
			RefreshToken: "1234",
			Expiry:       time.Time{},
		},
		AuthState: fmt.Sprintf("%v:%v", TestAdminUsername, TestAdminSecret),
	}
	testUser = &user.Model{
		ID:                fmt.Sprintf("%v@%v", TestUserName, TestSessionID),
		Username:          TestUserName,
		Secret:            TestUserSecret,
		SessionID:         TestSessionID,
		IsAdmin:           false,
		Score:             1,
		SpotifyAuthorized: false,
		AuthToken:         nil,
		AuthState:         fmt.Sprintf("%v:%v", TestUserName, TestUserSecret),
	}
)
