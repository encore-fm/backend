package config

import (
	"github.com/joho/godotenv"
	"github.com/prometheus/common/log"
	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

type SpotifyConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectUrl  string `mapstructure:"redirect_url"`
}

type ServerConfig struct {
	Port            int    `mapstructure:"port"`
	FrontendBaseUrl string `mapstructure:"frontend_base_url"`
	Debug           bool   `mapstructure:"debug"`
}

type DBConfig struct {
	DBUser                string `mapstructure:"db_user"`
	DBPassword            string `mapstructure:"db_password"`
	DBHost                string `mapstructure:"db_host"`
	DBPort                int    `mapstructure:"db_port"`
	DBName                string `mapstructure:"db_name"`
	UserCollectionName    string `mapstructure:"user_collection_name"`
	SessionCollectionName string `mapstructure:"session_collection_name"`
}

type Config struct {
	Spotify  *SpotifyConfig `mapstructure:"spotify"`
	Server   *ServerConfig  `mapstructure:"server"`
	Database *DBConfig      `mapstructure:"database"`
	MaxUsers int            `mapstructure:"max_users"`
}

var Conf *Config

// sets default configuration settings
func Setup() {
	err := godotenv.Load()
	if err != nil {
		log.Info("[startup] no .env file found")
	}

	// reading sensitive values from environment and adding them to config
	_ = viper.BindEnv("MONGO_HOST")
	mongoHost := viper.GetString("MONGO_HOST")

	_ = viper.BindEnv("MONGO_USER")
	mongoUser := viper.GetString("MONGO_USER")

	_ = viper.BindEnv("MONGO_PASSWORD")
	mongoPassword := viper.GetString("MONGO_PASSWORD")

	_ = viper.BindEnv("SPOTIFY_CLIENT_ID")
	clientID := viper.GetString("SPOTIFY_CLIENT_ID")

	_ = viper.BindEnv("SPOTIFY_CLIENT_SECRET")
	clientSecret := viper.GetString("SPOTIFY_CLIENT_SECRET")

	// heroku sets the PORT variable that you are supposed to bind
	_ = viper.BindEnv("PORT")
	port := viper.GetInt("PORT")

	// look if development flag is set
	_ = viper.BindEnv("DEVELOPMENT")
	development := viper.GetBool("DEVELOPMENT")

	// setup viper
	viper.SetConfigName("production")
	viper.SetConfigType("toml")
	viper.AddConfigPath("config/")
	if development {
		viper.SetConfigName("development")
	}

	c, err := FromFile()
	if err != nil {
		panic(err)
	}
	log.Infof("using config file at path %v", viper.ConfigFileUsed())

	Conf = c

	Conf.Database.DBHost = mongoHost
	Conf.Database.DBUser = mongoUser
	Conf.Database.DBPassword = mongoPassword
	Conf.Spotify.ClientID = clientID
	Conf.Spotify.ClientSecret = clientSecret
	Conf.Server.Port = port
	Conf.Server.Debug = development

	// setup logrus time stamps
	formatter := &logrus.TextFormatter{
		FullTimestamp: true,
	}
	logrus.SetFormatter(formatter)
}

// FromFile reads configuration from a file, bind a CLI flag to
func FromFile() (*Config, error) {
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := &Config{}
	if err := viper.Unmarshal(conf); err != nil {
		log.Errorf("unable to decode into config struct, %v", err)
	}

	return conf, nil
}
