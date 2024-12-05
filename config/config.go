package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config is the configuration for the application
type Config struct {
	TwitterAPIKey    string `envconfig:"TWITTER_API_KEY"`
	TwitterAPISecret string `envconfig:"TWITTER_API_SECRET"`
	TwitterUserID    string `envconfig:"TWITTER_USER_ID"`
	SecretKey        string `envconfig:"SECRET_KEY"`
	Port             string `envconfig:"PORT" default:"8080"`
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
