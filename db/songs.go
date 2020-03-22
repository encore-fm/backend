package db

import (
	"context"
	"errors"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/song"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SongCollection struct {
	collection *mongo.Collection
}

var (
	ErrSongAlreadyInQueue = errors.New("song already in queue")
)

func NewSongCollection(client *mongo.Client) *SongCollection {
	songCollection := client.
		Database(config.Conf.Database.DBName).
		Collection(config.Conf.Database.SongCollectionName)
	return &SongCollection{songCollection}
}

// GetSongByID returns a song struct if songID exists
// if songID does not exist it returns nil
func (h *SongCollection) GetSongByID(songID string) (*song.Model, error) {
	filter := bson.D{{"id", songID}}
	var foundSong *song.Model
	err := h.collection.FindOne(context.TODO(), filter).Decode(&foundSong)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return foundSong, nil
}

func (h *SongCollection) AddSong(newSong *song.Model) error {
	u, err := h.GetSongByID(newSong.ID)
	if err != nil {
		return err
	}
	if u != nil {
		return ErrSongAlreadyInQueue
	}

	_, err = h.collection.InsertOne(context.TODO(), newSong)
	return err
}