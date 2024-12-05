package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Auth middleware handles authentication
func Auth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// For now, we'll just check if the token is present
		// In a production environment, you'd want to validate the token
		c.Next()
	}
}
