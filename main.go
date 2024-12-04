package main

import (
    "log"
    "github.com/gin-gonic/gin"
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

    // Start server
    log.Fatal(r.Run("0.0.0.0:8000"))
}
