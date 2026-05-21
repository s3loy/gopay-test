package alipay

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
	"github.com/s3loy/gopay/internal/domain/entity"
	"github.com/s3loy/gopay/internal/domain/service"
	"github.com/s3loy/gopay/internal/pkg/apperror"
	"github.com/s3loy/gopay/internal/pkg/logger"
	"go.uber.org/zap"
)

type provider struct {
	client *Client
}

func NewProvider(client *Client) service.PaymentProvider {
	return &provider{client: client}
}

func (p *provider) Channel() entity.PaymentChannel {
	return entity.ChannelAlipay
}

func (p *provider) CreatePayment(ctx context.Context, req service.ProviderPaymentRequest) (*service.ProviderPaymentResult, error) {
	if !p.client.IsAvailable() {
		return nil, apperror.New(apperror.CodeAlipayAPICallFailed, "alipay client not available")
	}

	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", req.OrderNo).
		Set("total_amount", formatAmount(req.Amount)).
		Set("subject", req.Subject).
		Set("notify_url", req.NotifyURL)

	var result *service.ProviderPaymentResult

	switch req.Method {
	case entity.MethodPC, entity.MethodNative:
		aliRsp, err := p.client.V3().TradePrecreate(ctx, bm)
		if err != nil {
			logger.L().Error("alipay precreate failed", zap.Error(err))
			return nil, apperror.Wrap(err, apperror.CodeAlipayAPICallFailed)
		}
		if aliRsp.StatusCode != 200 {
			return nil, apperror.New(apperror.CodeAlipayAPICallFailed, aliRsp.ErrResponse.Message)
		}
		result = &service.ProviderPaymentResult{
			ThirdPartyNo: "",
			Status:       entity.PaymentStatusPending.String(),
			PayParams: map[string]any{
				"qr_code": aliRsp.QrCode,
			},
			RawResponse: map[string]any{"qr_code": aliRsp.QrCode},
		}

	case entity.MethodWAP, entity.MethodAPP:
		bm.Set("product_code", "QUICK_MSECURITY_PAY")
		aliRsp, err := p.client.V3().TradeCreate(ctx, bm)
		if err != nil {
			logger.L().Error("alipay trade create failed", zap.Error(err))
			return nil, apperror.Wrap(err, apperror.CodeAlipayAPICallFailed)
		}
		if aliRsp.StatusCode != 200 {
			return nil, apperror.New(apperror.CodeAlipayAPICallFailed, aliRsp.ErrResponse.Message)
		}
		result = &service.ProviderPaymentResult{
			ThirdPartyNo: aliRsp.TradeNo,
			Status:       entity.PaymentStatusPending.String(),
			PayParams: map[string]any{
				"trade_no": aliRsp.TradeNo,
			},
			RawResponse: map[string]any{"trade_no": aliRsp.TradeNo},
		}

	default:
		return nil, apperror.New(apperror.CodeInvalidPaymentMethod, fmt.Sprintf("unsupported alipay method: %s", req.Method))
	}

	return result, nil
}

func (p *provider) QueryPayment(ctx context.Context, thirdPartyNo string) (*service.ProviderPaymentResult, error) {
	if !p.client.IsAvailable() {
		return nil, apperror.New(apperror.CodeAlipayAPICallFailed, "alipay client not available")
	}

	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", thirdPartyNo)

	aliRsp, err := p.client.V3().TradeQuery(ctx, bm)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeAlipayAPICallFailed)
	}
	if aliRsp.StatusCode != 200 {
		return nil, apperror.New(apperror.CodeAlipayAPICallFailed, aliRsp.ErrResponse.Message)
	}

	return &service.ProviderPaymentResult{
		ThirdPartyNo: aliRsp.TradeNo,
		Status:       aliRsp.TradeStatus,
		RawResponse: map[string]any{
			"trade_status": aliRsp.TradeStatus,
		},
	}, nil
}

func (p *provider) Refund(ctx context.Context, req service.ProviderRefundRequest) (*service.ProviderRefundResult, error) {
	if !p.client.IsAvailable() {
		return nil, apperror.New(apperror.CodeAlipayAPICallFailed, "alipay client not available")
	}

	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", req.PaymentNo).
		Set("out_request_no", req.RefundNo).
		Set("refund_amount", formatAmount(req.Amount)).
		Set("refund_reason", req.Reason)

	aliRsp, err := p.client.V3().TradeRefund(ctx, bm)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeAlipayAPICallFailed)
	}
	if aliRsp.StatusCode != 200 {
		return nil, apperror.New(apperror.CodeAlipayAPICallFailed, aliRsp.ErrResponse.Message)
	}

	return &service.ProviderRefundResult{
		ThirdPartyNo: aliRsp.TradeNo,
		Status:       "SUCCESS",
		RawResponse: map[string]any{
			"refund_fee": aliRsp.RefundFee,
		},
	}, nil
}

func (p *provider) QueryRefund(ctx context.Context, thirdPartyNo string) (*service.ProviderRefundResult, error) {
	if !p.client.IsAvailable() {
		return nil, apperror.New(apperror.CodeAlipayAPICallFailed, "alipay client not available")
	}

	bm := make(gopay.BodyMap)
	bm.Set("out_request_no", thirdPartyNo)

	aliRsp, err := p.client.V3().TradeFastPayRefundQuery(ctx, bm)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeAlipayAPICallFailed)
	}
	if aliRsp.StatusCode != 200 {
		return nil, apperror.New(apperror.CodeAlipayAPICallFailed, aliRsp.ErrResponse.Message)
	}

	return &service.ProviderRefundResult{
		ThirdPartyNo: thirdPartyNo,
		Status:       aliRsp.RefundStatus,
		RawResponse: map[string]any{
			"refund_status": aliRsp.RefundStatus,
		},
	}, nil
}

func (p *provider) VerifyNotify(ctx context.Context, body []byte, headers map[string]string) (map[string]string, error) {
	if !p.client.IsAvailable() {
		return nil, apperror.New(apperror.CodeAlipayAPICallFailed, "alipay client not available")
	}

	publicKey := p.client.PublicKey()
	if publicKey == "" {
		return nil, apperror.New(apperror.CodeAlipayCertError, "alipay public key not configured")
	}

	values := make(url.Values)
	for k, v := range headers {
		values.Set(k, v)
	}

	notifyReq, err := alipay.ParseNotifyByURLValues(values)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeWebhookInvalidSignature)
	}

	ok, err := alipay.VerifySign(publicKey, notifyReq)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeWebhookInvalidSignature)
	}
	if !ok {
		return nil, apperror.New(apperror.CodeWebhookInvalidSignature, "alipay signature verification failed")
	}

	result := make(map[string]string)
	for k, v := range notifyReq {
		if s, ok := v.(string); ok {
			result[k] = s
		}
	}
	return result, nil
}

func ParseAlipayNotify(req *http.Request, publicKey string) (map[string]string, error) {
	if err := req.ParseForm(); err != nil {
		return nil, apperror.Wrap(err, apperror.CodeWebhookInvalidSignature)
	}

	notifyReq, err := alipay.ParseNotifyByURLValues(req.PostForm)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeWebhookInvalidSignature)
	}

	ok, err := alipay.VerifySign(publicKey, notifyReq)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeWebhookInvalidSignature)
	}
	if !ok {
		return nil, apperror.New(apperror.CodeWebhookInvalidSignature, "alipay signature verification failed")
	}

	result := make(map[string]string)
	for k, v := range notifyReq {
		if s, ok := v.(string); ok {
			result[k] = s
		}
	}
	return result, nil
}

func formatAmount(amount int64) string {
	return strconv.FormatFloat(float64(amount)/100, 'f', 2, 64)
}
