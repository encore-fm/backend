package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/antonbaumann/spotify-jukebox/config"
	"github.com/antonbaumann/spotify-jukebox/song"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SongCollection struct {
	collection *mongo.Collection
}

var (
	ErrSongAlreadyInQueue = errors.New("song already in queue")
	ErrUpdateNoSongWithID = errors.New("error updating song: no song with this ID")
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
	errMsg := "get song by id: %v"
	filter := bson.D{{"id", songID}}
	var foundSong *song.Model
	err := h.collection.FindOne(context.TODO(), filter).Decode(&foundSong)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	return foundSong, nil
}

func (h *SongCollection) AddSong(newSong *song.Model) error {
	errMsg := "ad song: %v"
	u, err := h.GetSongByID(newSong.ID)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if u != nil {
		return fmt.Errorf(errMsg, ErrSongAlreadyInQueue)
	}

	_, err = h.collection.InsertOne(context.TODO(), newSong)
	return fmt.Errorf(errMsg, err)
}

func (h *SongCollection) UpdateSong(updatedSong *song.Model) error {
	errMsg := "update song: %v"
	filter := bson.D{{"id", updatedSong.ID}}
	result, err := h.collection.ReplaceOne(context.TODO(), filter, updatedSong)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if result.ModifiedCount != 1 {
		return fmt.Errorf(errMsg, ErrUpdateNoSongWithID)
	}
	return nil
}

func (h *SongCollection) ListSongs() ([]*song.Model, error) {
	errMsg := "list songs: %v"
	opts := options.Find()
	opts.SetSort(bson.D{
		{"score", -1},
		{"time_added", 1},
	})

	cursor, err := h.collection.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	defer cursor.Close(context.TODO())

	var songList []*song.Model
	for cursor.Next(context.TODO()) {
		var elem song.Model
		if err := cursor.Decode(&elem); err != nil {
			return songList, fmt.Errorf(errMsg, err)
		}
		songList = append(songList, &elem)
	}
	return songList, nil
}
