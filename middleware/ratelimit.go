package middleware

import (
    "net/http"
    "sync"
    "time"
    "github.com/gin-gonic/gin"
    "twitter-bookmarks/config"
)

type RateLimiter struct {
    requests map[string][]time.Time
    mutex    sync.Mutex
}

var limiter = &RateLimiter{
    requests: make(map[string][]time.Time),
}

func RateLimit(config config.RateLimitConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientIP := c.ClientIP()
        
        limiter.mutex.Lock()
        defer limiter.mutex.Unlock()

        now := time.Now()
        
        // Clean old requests
        if requests, exists := limiter.requests[clientIP]; exists {
            var valid []time.Time
            for _, t := range requests {
                if now.Sub(t) <= config.TimeWindow {
                    valid = append(valid, t)
                }
            }
            limiter.requests[clientIP] = valid
        }

        // Check rate limit
        if len(limiter.requests[clientIP]) >= config.Requests {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
            })
            c.Abort()
            return
        }

        // Add new request
        limiter.requests[clientIP] = append(limiter.requests[clientIP], now)
        
        c.Next()
    }
}
