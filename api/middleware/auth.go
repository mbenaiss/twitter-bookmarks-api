package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TwitterTokenKey is the key for the Twitter token in the context
const TwitterTokenKey = "TWITTER_TOKEN"

type service interface {
	Authenticate(ctx context.Context) (string, error)
}

// Auth is a middleware to authenticate the user
func Auth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-API-KEY")
		if token != secretKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetTokenFromTwitter returns the token from the Twitter header
func GetTokenFromTwitter(service service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := service.Authenticate(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Set(TwitterTokenKey, token)
		c.Next()
	}
}
