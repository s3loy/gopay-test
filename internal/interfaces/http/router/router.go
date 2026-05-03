package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/s3loy/gopay/internal/interfaces/http/handler"
	"github.com/s3loy/gopay/internal/interfaces/http/middleware"
)

type Router struct {
	orderHandler   *handler.OrderHandler
	paymentHandler *handler.PaymentHandler
	refundHandler  *handler.RefundHandler
	webhookHandler *handler.WebhookHandler
	healthHandler  *handler.HealthHandler
}

func NewRouter(
	orderHandler *handler.OrderHandler,
	paymentHandler *handler.PaymentHandler,
	refundHandler *handler.RefundHandler,
	webhookHandler *handler.WebhookHandler,
	healthHandler *handler.HealthHandler,
) *Router {
	return &Router{
		orderHandler:   orderHandler,
		paymentHandler: paymentHandler,
		refundHandler:  refundHandler,
		webhookHandler: webhookHandler,
		healthHandler:  healthHandler,
	}
}

func (r *Router) Register(e *gin.Engine) {
	e.Use(middleware.Recovery())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORS())
	e.Use(middleware.Timeout(30 * time.Second))
	e.Use(middleware.Logger())

	e.GET("/health", r.healthHandler.Check)

	api := e.Group("/api/v1")
	{
		api.POST("/orders", r.orderHandler.Create)
		api.GET("/orders/:order_no", r.orderHandler.Get)
		api.POST("/orders/:order_no/close", r.orderHandler.Close)

		api.POST("/payments", r.paymentHandler.Create)
		api.GET("/payments/:payment_no", r.paymentHandler.Get)

		api.POST("/refunds", r.refundHandler.Create)
		api.GET("/refunds/:refund_no", r.refundHandler.Get)
	}

	webhook := e.Group("/webhook")
	{
		webhook.POST("/wechat/notify", r.webhookHandler.WechatNotify)
		webhook.POST("/alipay/notify", r.webhookHandler.AlipayNotify)
	}
}
