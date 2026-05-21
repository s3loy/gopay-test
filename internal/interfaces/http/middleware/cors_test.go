package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("allows configured origin", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
		c.Request.Header.Set("Origin", "http://localhost:3000")

		CORS([]string{"http://localhost:3000"})(c)

		if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
			t.Errorf("origin = %v, want http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("allows wildcard origin", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
		c.Request.Header.Set("Origin", "http://example.com")

		CORS([]string{"*"})(c)

		if w.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
			t.Errorf("origin = %v, want http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("blocks unconfigured origin", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
		c.Request.Header.Set("Origin", "http://evil.com")

		CORS([]string{"http://localhost:3000"})(c)

		if w.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Error("expected origin to be blocked")
		}
	})

	t.Run("allows all when no origins configured", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
		c.Request.Header.Set("Origin", "http://any.com")

		CORS([]string{})(c)

		if w.Header().Get("Access-Control-Allow-Origin") != "http://any.com" {
			t.Errorf("origin = %v, want http://any.com", w.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("handles OPTIONS preflight", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodOptions, "/test", nil)

		CORS([]string{"*"})(c)

		if w.Code != 204 {
			t.Errorf("status = %d, want 204", w.Code)
		}
	})

	t.Run("sets allowed methods and headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

		CORS([]string{"*"})(c)

		if w.Header().Get("Access-Control-Allow-Methods") == "" {
			t.Error("expected Access-Control-Allow-Methods to be set")
		}
		if w.Header().Get("Access-Control-Allow-Headers") == "" {
			t.Error("expected Access-Control-Allow-Headers to be set")
		}
	})
}
