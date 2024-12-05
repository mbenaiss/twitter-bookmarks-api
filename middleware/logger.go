package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger is a middleware that logs the request details
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		statusCode := c.Writer.Status()

		fmt.Printf("[%s] %s %s %d %s\n",
			end.Format("2006-01-02 15:04:05"),
			method,
			path,
			statusCode,
			latency,
		)
	}
}
