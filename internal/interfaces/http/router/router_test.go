package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/s3loy/gopay/internal/domain/entity"
	"github.com/s3loy/gopay/internal/interfaces/http/handler"
	"github.com/s3loy/gopay/internal/usecase"
)

func TestRouter_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderH := handler.NewOrderHandler(&mockRouterOrderUC{})
	paymentH := handler.NewPaymentHandler(&mockRouterPaymentUC{})
	refundH := handler.NewRefundHandler(&mockRouterRefundUC{})
	webhookH := handler.NewWebhookHandler(&mockRouterPaymentUC{})
	healthH := handler.NewHealthHandler()

	r := NewRouter(orderH, paymentH, refundH, webhookH, healthH, []string{"*"})
	engine := gin.New()
	r.Register(engine)

	tests := []struct {
		method string
		path   string
		body   string
		want   int
	}{
		{http.MethodGet, "/health", "", http.StatusOK},
		{http.MethodPost, "/api/v1/orders", `{"user_id":1,"subject":"test","amount":100}`, http.StatusOK},
		{http.MethodGet, "/api/v1/orders/ORD123", "", http.StatusOK},
		{http.MethodPost, "/api/v1/orders/ORD123/close", "", http.StatusOK},
		{http.MethodPost, "/api/v1/payments", `{"order_no":"ORD123","channel":"wechat","method":"native","client_ip":"127.0.0.1"}`, http.StatusOK},
		{http.MethodGet, "/api/v1/payments/PAY123", "", http.StatusOK},
		{http.MethodPost, "/api/v1/refunds", `{"payment_no":"PAY123","amount":100,"reason":"test"}`, http.StatusOK},
		{http.MethodGet, "/api/v1/refunds/REF123", "", http.StatusOK},
		{http.MethodPost, "/webhook/wechat/notify", "", http.StatusBadRequest},
		{http.MethodPost, "/webhook/alipay/notify", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			engine.ServeHTTP(w, req)

			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
		})
	}
}

// mock usecases for router test

type mockRouterOrderUC struct{}

func (m *mockRouterOrderUC) Create(ctx context.Context, userID uint64, subject string, amount int64, currency string, description string, expireMinutes int) (*entity.Order, error) {
	return &entity.Order{OrderNo: "ORD123", Status: entity.OrderStatusPending, ExpiredAt: time.Now().Add(time.Hour), CreatedAt: time.Now()}, nil
}
func (m *mockRouterOrderUC) Get(ctx context.Context, orderNo string) (*entity.Order, error) {
	return &entity.Order{OrderNo: orderNo, Status: entity.OrderStatusPending, ExpiredAt: time.Now().Add(time.Hour), CreatedAt: time.Now()}, nil
}
func (m *mockRouterOrderUC) Close(ctx context.Context, orderNo string) error {
	return nil
}

type mockRouterPaymentUC struct{}

func (m *mockRouterPaymentUC) Create(ctx context.Context, req usecase.CreatePaymentRequest) (*entity.Payment, map[string]any, error) {
	return &entity.Payment{PaymentNo: "PAY123", Status: entity.PaymentStatusPending, ExpireAt: time.Now().Add(time.Hour), CreatedAt: time.Now()}, nil, nil
}
func (m *mockRouterPaymentUC) Get(ctx context.Context, paymentNo string) (*entity.Payment, error) {
	return &entity.Payment{PaymentNo: paymentNo, Status: entity.PaymentStatusPending, ExpireAt: time.Now().Add(time.Hour), CreatedAt: time.Now()}, nil
}
func (m *mockRouterPaymentUC) HandleWechatNotify(ctx context.Context, body []byte, headers map[string]string) error {
	return nil
}
func (m *mockRouterPaymentUC) HandleAlipayNotify(ctx context.Context, params map[string]string) error {
	return nil
}

type mockRouterRefundUC struct{}

func (m *mockRouterRefundUC) Create(ctx context.Context, req usecase.CreateRefundRequest) (*entity.Refund, error) {
	return &entity.Refund{RefundNo: "REF123", Status: entity.RefundStatusSuccess, CreatedAt: time.Now()}, nil
}
func (m *mockRouterRefundUC) Get(ctx context.Context, refundNo string) (*entity.Refund, error) {
	return &entity.Refund{RefundNo: refundNo, Status: entity.RefundStatusSuccess, CreatedAt: time.Now()}, nil
}
