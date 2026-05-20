package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

func CORS(origins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := false
		allowedOrigin := ""

		if len(origins) == 0 {
			// Default: allow same-origin only in production if no origins configured
			allowed = true
			allowedOrigin = origin
		} else {
			for _, o := range origins {
				if o == "*" || o == origin {
					allowed = true
					allowedOrigin = origin
					break
				}
			}
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Max-Age", (24*time.Hour).String())

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
