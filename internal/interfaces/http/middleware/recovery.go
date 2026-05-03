package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/s3loy/gopay/internal/pkg/logger"
	"go.uber.org/zap"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		logger.L().Error("panic recovered", zap.Any("error", err), zap.String("path", c.Request.URL.Path))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code":    10002,
			"message": "internal server error",
		})
	})
}
