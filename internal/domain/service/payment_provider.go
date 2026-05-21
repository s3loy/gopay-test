package service

import (
	"context"

	"github.com/s3loy/gopay/internal/domain/entity"
)

type ProviderPaymentRequest struct {
	OrderNo       string
	Subject       string
	Amount        int64
	Currency      string
	Method        entity.PaymentMethod
	ClientIP      string
	NotifyURL     string
	ReturnURL     string
	OpenID        string // WeChat JSAPI
	BuyerID       string // Alipay WAP
	ExpireMinutes int
}

type ProviderPaymentResult struct {
	ThirdPartyNo string
	Status       string
	PayParams    map[string]any // 渠道特定参数，如二维码URL、JSAPI调起参数等
	RawResponse  map[string]any
}

type ProviderRefundRequest struct {
	PaymentNo    string
	ThirdPartyNo string
	RefundNo     string
	Amount       int64
	Reason       string
	NotifyURL    string
}

type ProviderRefundResult struct {
	ThirdPartyNo string
	Status       string
	RawResponse  map[string]any
}

type PaymentProvider interface {
	CreatePayment(ctx context.Context, req ProviderPaymentRequest) (*ProviderPaymentResult, error)
	QueryPayment(ctx context.Context, thirdPartyNo string) (*ProviderPaymentResult, error)
	Refund(ctx context.Context, req ProviderRefundRequest) (*ProviderRefundResult, error)
	QueryRefund(ctx context.Context, thirdPartyNo string) (*ProviderRefundResult, error)
	VerifyNotify(ctx context.Context, body []byte, headers map[string]string) (map[string]string, error)
	Channel() entity.PaymentChannel
}

type PaymentProviderFactory interface {
	Get(channel entity.PaymentChannel) (PaymentProvider, error)
}
