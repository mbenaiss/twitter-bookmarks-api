package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config is the configuration for the application
type Config struct {
	TwitterConsumerKey    string `envconfig:"TWITTER_CONSUMER_KEY"`
	TwitterConsumerSecret string `envconfig:"TWITTER_CONSUMER_SECRET"`
	TwitterAccessToken    string `envconfig:"TWITTER_ACCESS_TOKEN"`
	TwitterAccessSecret   string `envconfig:"TWITTER_ACCESS_SECRET"`
	TwitterUserID         string `envconfig:"TWITTER_USER_ID"`
	SecretKey             string `envconfig:"SECRET_KEY"`
	TwitterClientID       string `envconfig:"TWITTER_CLIENT_ID"`
	TwitterClientSecret   string `envconfig:"TWITTER_CLIENT_SECRET"`
	TwitterAuthToken      string `envconfig:"TWITTER_AUTH_TOKEN"`
	TwitterRedirectURI    string `envconfig:"TWITTER_REDIRECT_URI"`
	Port                  string `envconfig:"PORT" default:"8080"`
}

// Load loads the configuration from the environment variables
func Load() (Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found")
	}

	var c Config
	err = envconfig.Process("", &c)
	if err != nil {
		return Config{}, fmt.Errorf("unable to get envconfig %w", err)
	}

	return c, nil
}
