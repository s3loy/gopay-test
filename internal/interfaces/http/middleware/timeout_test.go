package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("request completes within timeout", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

		r.Use(Timeout(5 * time.Second))
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		r.HandleContext(c)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}
