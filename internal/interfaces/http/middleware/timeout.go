package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/s3loy/gopay/internal/pkg/response"
)

func Timeout(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			c.Abort()
			if !c.Writer.Written() {
				response.ErrorWithCode(c, 10004)
			}
		}
	}
}
