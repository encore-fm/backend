package spotifycl

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const RefreshWaitTime = time.Minute * 50

type SpotifyClient struct {
	Client spotify.Client
	config *clientcredentials.Config
	ticker *time.Ticker
	quit   chan struct{}
}

func New(clientID, clientSecret string) (*SpotifyClient, error) {
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     spotify.TokenURL,
	}
	token, err := config.Token(context.TODO())
	if err != nil {
		return nil, err
	}
	client := spotify.Authenticator{}.NewClient(token)

	return &SpotifyClient{
		Client: client,
		config: config,
		ticker: time.NewTicker(RefreshWaitTime),
		quit:   make(chan struct{}),
	}, nil
}

func (c *SpotifyClient) Start() {
	go func() {
		for {
			select {
			case <-c.ticker.C:
				c.refreshToken()
			case <-c.quit:
				c.ticker.Stop()
				return
			}
		}
	}()
}

func (c *SpotifyClient) refreshToken() {
	token, err := c.config.Token(context.TODO())
	if err != nil {
		log.Errorf("[spotifycl] refreshing token: %v", err)
	}
	c.Client = spotify.Authenticator{}.NewClient(token)
	log.Info("[spotifycl] refreshed token")
}

func (c *SpotifyClient) GetClientToken() (*oauth2.Token, error){
	c.refreshToken()
	return c.Client.Token()
}
