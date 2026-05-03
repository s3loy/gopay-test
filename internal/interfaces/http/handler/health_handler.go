package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/s3loy/gopay/internal/pkg/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok"})
}
