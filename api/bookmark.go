package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"twitter-bookmarks/api/middleware"
	"twitter-bookmarks/models"
)

type service interface {
	Authenticate(ctx context.Context) (string, error)
	GetBookmarks(ctx context.Context, token string) (*models.BookmarkResponse, error)
	GetBookmarksAfterDate(ctx context.Context, afterDate time.Time, token string) (*models.BookmarkResponse, error)
}

func (s *Server) getBookmarks(service service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := c.Get(middleware.TwitterTokenKey)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		response, err := service.GetBookmarks(c.Request.Context(), token.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch bookmarks",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func (s *Server) getBookmarksWithDateFilter(service service) gin.HandlerFunc {
	return func(c *gin.Context) {
		dateStr := c.Query("after")
		afterDate, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid date format. Use RFC3339",
			})
			return
		}

		token, ok := c.Get(middleware.TwitterTokenKey)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		response, err := service.GetBookmarksAfterDate(c.Request.Context(), afterDate, token.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch bookmarks",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}
