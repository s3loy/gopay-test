package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("generates new request id", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

		RequestID()(c)

		rid := c.Writer.Header().Get(HeaderRequestID)
		if rid == "" {
			t.Error("expected request id to be generated")
		}
		if c.GetString("request_id") != rid {
			t.Error("request_id in context should match header")
		}
	})

	t.Run("uses existing request id from header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
		c.Request.Header.Set(HeaderRequestID, "req-123")

		RequestID()(c)

		if c.Writer.Header().Get(HeaderRequestID) != "req-123" {
			t.Errorf("expected request id to be req-123, got %s", c.Writer.Header().Get(HeaderRequestID))
		}
	})
}
