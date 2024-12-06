package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"twitter-bookmarks/api/middleware"
	"twitter-bookmarks/models"
)

type service interface {
	Authenticate(ctx context.Context, code string) (string, error)
	GetBookmarks(ctx context.Context, token string) (*models.BookmarkResponse, error)
	GetBookmarksAfterDate(ctx context.Context, token string, date time.Time) (*models.BookmarkResponse, error)
}

func (s *Server) authenticate(service service, codeVerifier string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := service.Authenticate(c.Request.Context(), codeVerifier)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to authenticate",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

func (s *Server) getBookmarks(service service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := c.Get(middleware.TwitterTokenKey)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		fmt.Println(token)

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

		response, err := service.GetBookmarksAfterDate(c.Request.Context(), token.(string), afterDate)
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
