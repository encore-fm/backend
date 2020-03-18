package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type SpotifyConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DBConfig struct {
	DBHost             string `mapstructure:"db_host"`
	DBPort             int    `mapstructure:"db_port"`
	DBName             string `mapstructure:"db_name"`
	UserCollectionName string `mapstructure:"user_collection_name"`
	SongCollectionName string `mapstructure:"song_collection_name"`
}

type Config struct {
	Spotify  *SpotifyConfig `mapstructure:"spotify"`
	Server   *ServerConfig  `mapstructure:"server"`
	Database *DBConfig      `mapstructure:"database"`
	MaxUsers int            `mapstructure:"max_users"`
}

var Conf *Config

// init sets default configuration file settings such as
// path look up values
func init() {
	// Config file lookup locations
	viper.SetConfigType("toml")
	viper.SetConfigName("spotify-jukebox")
	viper.AddConfigPath("$HOME/.config/")
	viper.AddConfigPath("config/") // this is only the example file with dummy values

	c, err := FromFile()
	if err != nil {
		panic(err)
	}
	Conf = c
}

// FromFile reads configuration from a file, bind a CLI flag to
func FromFile() (*Config, error) {
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := &Config{}
	if err := viper.Unmarshal(conf); err != nil {
		fmt.Printf("unable to decode into config struct, %v", err)
	}

	return conf, nil
}
