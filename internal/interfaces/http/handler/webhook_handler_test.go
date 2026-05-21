package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/s3loy/gopay/internal/domain/entity"
	"github.com/s3loy/gopay/internal/pkg/apperror"
	"github.com/s3loy/gopay/internal/usecase"
)

type mockWebhookPaymentUC struct {
	err error
}

func (m *mockWebhookPaymentUC) Create(ctx context.Context, req usecase.CreatePaymentRequest) (*entity.Payment, map[string]any, error) {
	return nil, nil, nil
}
func (m *mockWebhookPaymentUC) Get(ctx context.Context, paymentNo string) (*entity.Payment, error) {
	return nil, nil
}
func (m *mockWebhookPaymentUC) HandleWechatNotify(ctx context.Context, body []byte, headers map[string]string) error {
	return m.err
}
func (m *mockWebhookPaymentUC) HandleAlipayNotify(ctx context.Context, params map[string]string) error {
	return m.err
}

func setupWebhookRouter(h *WebhookHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/webhook/wechat/notify", h.WechatNotify)
	r.POST("/webhook/alipay/notify", h.AlipayNotify)
	return r
}

func TestWebhookHandler_WechatNotify(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h := NewWebhookHandler(&mockWebhookPaymentUC{})
		r := setupWebhookRouter(h)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/webhook/wechat/notify", bytes.NewBufferString("{}"))
		req.Header.Set("Wechatpay-Signature", "sig")
		req.Header.Set("Wechatpay-Serial", "serial")
		req.Header.Set("Wechatpay-Nonce", "nonce")
		req.Header.Set("Wechatpay-Timestamp", "1234567890")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
		if !strings.Contains(w.Body.String(), "success") {
			t.Errorf("body = %s, want success", w.Body.String())
		}
	})

	t.Run("missing header", func(t *testing.T) {
		h := NewWebhookHandler(&mockWebhookPaymentUC{})
		r := setupWebhookRouter(h)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/webhook/wechat/notify", bytes.NewBufferString("{}"))
		req.Header.Set("Wechatpay-Signature", "sig")
		req.Header.Set("Wechatpay-Serial", "serial")
		req.Header.Set("Wechatpay-Nonce", "nonce")
		// missing Wechatpay-Timestamp
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("handler error", func(t *testing.T) {
		h := NewWebhookHandler(&mockWebhookPaymentUC{err: apperror.New(apperror.CodePaymentNotFound, "not found")})
		r := setupWebhookRouter(h)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/webhook/wechat/notify", bytes.NewBufferString("{}"))
		req.Header.Set("Wechatpay-Signature", "sig")
		req.Header.Set("Wechatpay-Serial", "serial")
		req.Header.Set("Wechatpay-Nonce", "nonce")
		req.Header.Set("Wechatpay-Timestamp", "1234567890")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestWebhookHandler_AlipayNotify(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h := NewWebhookHandler(&mockWebhookPaymentUC{})
		r := setupWebhookRouter(h)

		body := "out_trade_no=PAY123&trade_status=TRADE_SUCCESS"
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/webhook/alipay/notify", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("missing out_trade_no", func(t *testing.T) {
		h := NewWebhookHandler(&mockWebhookPaymentUC{})
		r := setupWebhookRouter(h)

		body := "trade_status=TRADE_SUCCESS"
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/webhook/alipay/notify", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("handler error", func(t *testing.T) {
		h := NewWebhookHandler(&mockWebhookPaymentUC{err: apperror.New(apperror.CodePaymentNotFound, "not found")})
		r := setupWebhookRouter(h)

		body := "out_trade_no=PAY123&trade_status=TRADE_SUCCESS"
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/webhook/alipay/notify", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}
