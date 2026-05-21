package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/s3loy/gopay/internal/pkg/logger"
	"github.com/s3loy/gopay/internal/usecase"
	"go.uber.org/zap"
)

const maxWebhookBodySize = 1 << 20 // 1MB

type WebhookHandler struct {
	paymentUC usecase.PaymentUsecase
}

func NewWebhookHandler(paymentUC usecase.PaymentUsecase) *WebhookHandler {
	return &WebhookHandler{paymentUC: paymentUC}
}

func (h *WebhookHandler) WechatNotify(c *gin.Context) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxWebhookBodySize))
	if err != nil {
		logger.L().Error("read wechat notify body failed", zap.Error(err))
		c.String(http.StatusBadRequest, "fail")
		return
	}

	requiredHeaders := []string{"Wechatpay-Signature", "Wechatpay-Serial", "Wechatpay-Nonce", "Wechatpay-Timestamp"}
	for _, name := range requiredHeaders {
		if c.GetHeader(name) == "" {
			logger.L().Error("missing required wechat notify header", zap.String("header", name))
			c.String(http.StatusBadRequest, "fail")
			return
		}
	}

	headers := map[string]string{
		"Wechatpay-Signature": c.GetHeader("Wechatpay-Signature"),
		"Wechatpay-Serial":    c.GetHeader("Wechatpay-Serial"),
		"Wechatpay-Nonce":     c.GetHeader("Wechatpay-Nonce"),
		"Wechatpay-Timestamp": c.GetHeader("Wechatpay-Timestamp"),
	}

	if err := h.paymentUC.HandleWechatNotify(c.Request.Context(), body, headers); err != nil {
		logger.L().Error("handle wechat notify failed", zap.Error(err), zap.String("out_trade_no", headers["Wechatpay-Nonce"]))
		c.String(http.StatusBadRequest, "fail")
		return
	}

	c.String(http.StatusOK, "success")
}

func (h *WebhookHandler) AlipayNotify(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		logger.L().Error("parse alipay notify form failed", zap.Error(err))
		c.String(http.StatusBadRequest, "fail")
		return
	}

	params := make(map[string]string)
	for k, v := range c.Request.PostForm {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	if params["out_trade_no"] == "" {
		logger.L().Error("missing required alipay notify param", zap.String("param", "out_trade_no"))
		c.String(http.StatusBadRequest, "fail")
		return
	}

	if err := h.paymentUC.HandleAlipayNotify(c.Request.Context(), params); err != nil {
		logger.L().Error("handle alipay notify failed", zap.Error(err), zap.String("out_trade_no", params["out_trade_no"]))
		c.String(http.StatusBadRequest, "fail")
		return
	}

	c.String(http.StatusOK, "success")
}
