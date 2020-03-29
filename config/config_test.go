package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromFile(t *testing.T) {
	result, err := FromFile()

	assert.Nil(t, err)
	assert.Equal(t, 1000, result.MaxUsers)

	// test admin config
	adminConfig := &AdminConfig{
		Username: "admin",
		Password: "password",
	}
	assert.Equal(t, adminConfig, result.Admin)

	// test spotify config
	spotifyConfig := &SpotifyConfig{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectUrl:  "redirect.url",
		State:        "state",
		OpenBrowser:  true,
	}
	assert.Equal(t, spotifyConfig, result.Spotify)

	// test server config
	serverConfig := &ServerConfig{
		Port:            8080,
		FrontendBaseUrl: "http://localhost:3000",
	}
	assert.Equal(t, serverConfig, result.Server)

	// test database config
	dbConfig := &DBConfig{
		DBUser:                "root",
		DBPassword:            "root",
		DBHost:                "127.0.0.1",
		DBPort:                27017,
		DBName:                "spotify-jukebox",
		UserCollectionName:    "users",
		SongCollectionName:    "songs",
		SessionCollectionName: "sessions",
	}
	assert.Equal(t, dbConfig, result.Database)
}
