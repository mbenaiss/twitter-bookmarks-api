package config

import (
    "os"
    "time"
)

type Config struct {
    TwitterAPIKey      string
    TwitterAPISecret   string
    JWTSecret         string
    RateLimit         RateLimitConfig
}

type RateLimitConfig struct {
    Requests    int
    TimeWindow  time.Duration
}

func Load() *Config {
    return &Config{
        TwitterAPIKey:    getEnvOrDefault("TWITTER_API_KEY", ""),
        TwitterAPISecret: getEnvOrDefault("TWITTER_API_SECRET", ""),
        JWTSecret:       getEnvOrDefault("JWT_SECRET", "your-secret-key"),
        RateLimit: RateLimitConfig{
            Requests:   100,
            TimeWindow: time.Minute,
        },
    }
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
