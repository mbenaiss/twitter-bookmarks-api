package main

import (
    "log"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt"
    "twitter-bookmarks/config"
    "twitter-bookmarks/handlers"
    "twitter-bookmarks/middleware"
)

func main() {
    // Initialize router
    r := gin.Default()

    // Load configuration
    cfg := config.Load()

    // Apply global middleware
    r.Use(gin.Logger())
    r.Use(middleware.RateLimit(cfg.RateLimit))
    r.Use(middleware.CORS())

    // API routes
    api := r.Group("/api/v1")
    api.Use(middleware.Auth())

    // Bookmark endpoints
    api.GET("/bookmarks", handlers.GetBookmarks)
    api.GET("/bookmarks/filter", handlers.GetBookmarksWithDateFilter)

    // Health check
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // Development only - token generator
    if gin.Mode() != gin.ReleaseMode {
        r.GET("/debug/token", func(c *gin.Context) {
            token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
                "user_id": "12345", // Test user ID
                "exp":     time.Now().Add(time.Hour * 24).Unix(),
            })
            tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
            if err != nil {
                c.JSON(500, gin.H{"error": "Could not generate token"})
                return
            }
            c.JSON(200, gin.H{"token": tokenString})
        })
    }

    // Start server
    log.Fatal(r.Run("0.0.0.0:8000"))
}
