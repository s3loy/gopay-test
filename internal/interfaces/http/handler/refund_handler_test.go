package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/s3loy/gopay/internal/domain/entity"
	"github.com/s3loy/gopay/internal/pkg/apperror"
	"github.com/s3loy/gopay/internal/usecase"
)

type mockRefundUC struct {
	refund *entity.Refund
	err    error
}

func (m *mockRefundUC) Create(ctx context.Context, req usecase.CreateRefundRequest) (*entity.Refund, error) {
	return m.refund, m.err
}
func (m *mockRefundUC) Get(ctx context.Context, refundNo string) (*entity.Refund, error) {
	return m.refund, m.err
}

func setupRefundRouter(h *RefundHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/refunds", h.Create)
	r.GET("/refunds/:refund_no", h.Get)
	return r
}

func TestRefundHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		refund := &entity.Refund{RefundNo: "REF123", PaymentNo: "PAY123", OrderNo: "ORD123", Amount: 100, Reason: "test", Status: entity.RefundStatusSuccess, CreatedAt: time.Now()}
		h := NewRefundHandler(&mockRefundUC{refund: refund})
		r := setupRefundRouter(h)

		body := `{"payment_no":"PAY123","amount":100,"reason":"test"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/refunds", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		h := NewRefundHandler(&mockRefundUC{})
		r := setupRefundRouter(h)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/refunds", bytes.NewBufferString("{invalid"))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		h := NewRefundHandler(&mockRefundUC{})
		r := setupRefundRouter(h)

		body := `{"payment_no":"","amount":0,"reason":""}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/refunds", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("usecase error", func(t *testing.T) {
		h := NewRefundHandler(&mockRefundUC{err: apperror.New(apperror.CodePaymentNotFound, "not found")})
		r := setupRefundRouter(h)

		body := `{"payment_no":"PAY123","amount":100,"reason":"test"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/refunds", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestRefundHandler_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		refund := &entity.Refund{RefundNo: "REF123", Status: entity.RefundStatusSuccess, CreatedAt: time.Now()}
		h := NewRefundHandler(&mockRefundUC{refund: refund})
		r := setupRefundRouter(h)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/refunds/REF123", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("not found", func(t *testing.T) {
		h := NewRefundHandler(&mockRefundUC{err: apperror.New(apperror.CodeRefundNotFound, "not found")})
		r := setupRefundRouter(h)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/refunds/REF999", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
