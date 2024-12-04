package handlers

import (
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
    "twitter-bookmarks/services"
    "twitter-bookmarks/models"
)

func GetBookmarks(c *gin.Context) {
    userID := c.GetString("user_id")
    
    twitterService := services.NewTwitterService()
    var response *models.BookmarkResponse
    response, err := twitterService.GetBookmarks(userID)
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to fetch bookmarks",
        })
        return
    }

    c.JSON(http.StatusOK, response)
}

func GetBookmarksWithDateFilter(c *gin.Context) {
    userID := c.GetString("user_id")
    dateStr := c.Query("after")
    
    afterDate, err := time.Parse(time.RFC3339, dateStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid date format. Use RFC3339",
        })
        return
    }

    twitterService := services.NewTwitterService()
    var response *models.BookmarkResponse
    response, err = twitterService.GetBookmarksAfterDate(userID, afterDate)
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to fetch bookmarks",
        })
        return
    }

    c.JSON(http.StatusOK, response)
}
