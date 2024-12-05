package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
