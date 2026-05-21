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

type mockPaymentUC struct {
	payment   *entity.Payment
	payParams map[string]any
	err       error
}

func (m *mockPaymentUC) Create(ctx context.Context, req usecase.CreatePaymentRequest) (*entity.Payment, map[string]any, error) {
	return m.payment, m.payParams, m.err
}
func (m *mockPaymentUC) Get(ctx context.Context, paymentNo string) (*entity.Payment, error) {
	return m.payment, m.err
}
func (m *mockPaymentUC) HandleWechatNotify(ctx context.Context, body []byte, headers map[string]string) error {
	return m.err
}
func (m *mockPaymentUC) HandleAlipayNotify(ctx context.Context, params map[string]string) error {
	return m.err
}

func setupPaymentRouter(h *PaymentHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/payments", h.Create)
	r.GET("/payments/:payment_no", h.Get)
	return r
}

func TestPaymentHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		payment := &entity.Payment{PaymentNo: "PAY123", OrderNo: "ORD123", Amount: 100, Status: entity.PaymentStatusPending, ExpireAt: time.Now().Add(time.Hour), CreatedAt: time.Now()}
		h := NewPaymentHandler(&mockPaymentUC{payment: payment, payParams: map[string]any{"qr_code": "https://qr.test"}})
		r := setupPaymentRouter(h)

		body := `{"order_no":"ORD123","channel":"wechat","method":"native","client_ip":"127.0.0.1"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		h := NewPaymentHandler(&mockPaymentUC{})
		r := setupPaymentRouter(h)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewBufferString("{invalid"))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		h := NewPaymentHandler(&mockPaymentUC{})
		r := setupPaymentRouter(h)

		body := `{"order_no":"","channel":"invalid","method":"invalid","client_ip":"invalid"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("usecase error", func(t *testing.T) {
		h := NewPaymentHandler(&mockPaymentUC{err: apperror.New(apperror.CodeOrderNotFound, "not found")})
		r := setupPaymentRouter(h)

		body := `{"order_no":"ORD123","channel":"wechat","method":"native","client_ip":"127.0.0.1"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestPaymentHandler_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		payment := &entity.Payment{PaymentNo: "PAY123", Status: entity.PaymentStatusPending, ExpireAt: time.Now().Add(time.Hour), CreatedAt: time.Now()}
		h := NewPaymentHandler(&mockPaymentUC{payment: payment})
		r := setupPaymentRouter(h)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/PAY123", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("not found", func(t *testing.T) {
		h := NewPaymentHandler(&mockPaymentUC{err: apperror.New(apperror.CodePaymentNotFound, "not found")})
		r := setupPaymentRouter(h)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/payments/PAY999", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestToPaymentResponse(t *testing.T) {
	now := time.Now()
	payment := &entity.Payment{
		PaymentNo: "PAY123",
		OrderNo:   "ORD123",
		Channel:   entity.ChannelWechat,
		Method:    entity.MethodNative,
		Amount:    100,
		Currency:  "CNY",
		Status:    entity.PaymentStatusPending,
		ExpireAt:  now,
		CreatedAt: now,
	}
	resp := toPaymentResponse(payment)
	if resp.PaymentNo != "PAY123" {
		t.Errorf("PaymentNo = %s, want PAY123", resp.PaymentNo)
	}
	if resp.Channel != "wechat" {
		t.Errorf("Channel = %s, want wechat", resp.Channel)
	}
}
