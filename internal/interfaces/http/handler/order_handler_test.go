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

type mockOrderUC struct {
	order *entity.Order
	err   error
}

var _ usecase.OrderUsecase = (*mockOrderUC)(nil)

func (m *mockOrderUC) Create(ctx context.Context, userID uint64, subject string, amount int64, currency string, description string, expireMinutes int) (*entity.Order, error) {
	return m.order, m.err
}
func (m *mockOrderUC) Get(ctx context.Context, orderNo string) (*entity.Order, error) {
	return m.order, m.err
}
func (m *mockOrderUC) Close(ctx context.Context, orderNo string) error {
	return m.err
}

func setupOrderRouter(h *OrderHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/orders", h.Create)
	r.GET("/orders/:order_no", h.Get)
	r.POST("/orders/:order_no/close", h.Close)
	return r
}

func TestOrderHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		order := &entity.Order{OrderNo: "ORD123", Amount: 100, Currency: "CNY", Status: entity.OrderStatusPending, ExpiredAt: time.Now().Add(time.Hour), CreatedAt: time.Now()}
		h := NewOrderHandler(&mockOrderUC{order: order})
		r := setupOrderRouter(h)

		body := `{"user_id":1,"subject":"test","amount":100}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		h := NewOrderHandler(&mockOrderUC{})
		r := setupOrderRouter(h)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString("{invalid"))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		h := NewOrderHandler(&mockOrderUC{})
		r := setupOrderRouter(h)

		body := `{"user_id":0,"subject":"","amount":0}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("usecase error", func(t *testing.T) {
		h := NewOrderHandler(&mockOrderUC{err: apperror.New(apperror.CodeOrderAmountInvalid, "invalid")})
		r := setupOrderRouter(h)

		body := `{"user_id":1,"subject":"test","amount":100}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestOrderHandler_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		order := &entity.Order{OrderNo: "ORD123", Status: entity.OrderStatusPending, ExpiredAt: time.Now().Add(time.Hour), CreatedAt: time.Now()}
		h := NewOrderHandler(&mockOrderUC{order: order})
		r := setupOrderRouter(h)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/orders/ORD123", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("not found", func(t *testing.T) {
		h := NewOrderHandler(&mockOrderUC{err: apperror.New(apperror.CodeOrderNotFound, "not found")})
		r := setupOrderRouter(h)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/orders/ORD999", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestOrderHandler_Close(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h := NewOrderHandler(&mockOrderUC{})
		r := setupOrderRouter(h)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/orders/ORD123/close", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("close error", func(t *testing.T) {
		h := NewOrderHandler(&mockOrderUC{err: apperror.New(apperror.CodeOrderAlreadyClosed, "already closed")})
		r := setupOrderRouter(h)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/orders/ORD123/close", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
		}
	})
}

func TestToOrderResponse(t *testing.T) {
	now := time.Now()
	order := &entity.Order{
		OrderNo:   "ORD123",
		UserID:    1,
		Subject:   "test",
		Amount:    100,
		Currency:  "CNY",
		Status:    entity.OrderStatusPending,
		ExpiredAt: now,
		CreatedAt: now,
	}
	resp := toOrderResponse(order)
	if resp.OrderNo != "ORD123" {
		t.Errorf("OrderNo = %s, want ORD123", resp.OrderNo)
	}
	if resp.PaidAt != nil {
		t.Error("PaidAt should be nil")
	}

	// Test with PaidAt set
	order.PaidAt = &now
	resp = toOrderResponse(order)
	if resp.PaidAt == nil {
		t.Error("PaidAt should not be nil")
	}
}
