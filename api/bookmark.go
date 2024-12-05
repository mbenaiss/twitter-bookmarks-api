package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"twitter-bookmarks/models"
)

type service interface {
	Authenticate(ctx context.Context) (string, error)
	GetBookmarks(ctx context.Context, userID string, token string) (*models.BookmarkResponse, error)
	GetBookmarksAfterDate(ctx context.Context, userID string, afterDate time.Time, token string) (*models.BookmarkResponse, error)
}

func (s *Server) getBookmarks(service service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		token := c.GetHeader("Authorization")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "User ID is required",
				"code":  "MISSING_USER_ID",
			})
			return
		}

		response, err := service.GetBookmarks(c.Request.Context(), userID, token)
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
		userID := c.Param("user_id")
		dateStr := c.Query("after")
		token := c.GetHeader("Authorization")
		afterDate, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid date format. Use RFC3339",
			})
			return
		}

		response, err := service.GetBookmarksAfterDate(c.Request.Context(), userID, afterDate, token)
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
