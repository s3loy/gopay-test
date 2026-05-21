package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-pay/gopay"
	wechatv3 "github.com/go-pay/gopay/wechat/v3"
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
	return entity.ChannelWechat
}

func (p *provider) CreatePayment(ctx context.Context, req service.ProviderPaymentRequest) (*service.ProviderPaymentResult, error) {
	if !p.client.IsAvailable() {
		return nil, apperror.New(apperror.CodeWeChatAPICallFailed, "wechat client not available")
	}

	appID := p.client.AppID()
	if appID == "" {
		return nil, apperror.New(apperror.CodeWeChatAPICallFailed, "wechat appid not configured")
	}

	bm := make(gopay.BodyMap)
	bm.Set("appid", appID).
		Set("description", req.Subject).
		Set("out_trade_no", req.OrderNo).
		Set("notify_url", req.NotifyURL).
		SetBodyMap("amount", func(bm gopay.BodyMap) {
			bm.Set("total", req.Amount).Set("currency", req.Currency)
		})

	var result *service.ProviderPaymentResult

	switch req.Method {
	case entity.MethodNative:
		wxRsp, err := p.client.V3().V3TransactionNative(ctx, bm)
		if err != nil {
			logger.L().Error("wechat native payment failed", zap.Error(err))
			return nil, apperror.Wrap(err, apperror.CodeWeChatAPICallFailed)
		}
		if wxRsp.Code != wechatv3.Success {
			return nil, apperror.New(apperror.CodeWeChatAPICallFailed, wxRsp.Error)
		}
		result = &service.ProviderPaymentResult{
			ThirdPartyNo: "",
			Status:       entity.PaymentStatusPending.String(),
			PayParams: map[string]any{
				"code_url": wxRsp.Response.CodeUrl,
			},
			RawResponse: map[string]any{"code_url": wxRsp.Response.CodeUrl},
		}

	case entity.MethodJSAPI:
		bm.SetBodyMap("payer", func(bm gopay.BodyMap) {
			bm.Set("openid", req.OpenID)
		})
		wxRsp, err := p.client.V3().V3TransactionJsapi(ctx, bm)
		if err != nil {
			logger.L().Error("wechat jsapi payment failed", zap.Error(err))
			return nil, apperror.Wrap(err, apperror.CodeWeChatAPICallFailed)
		}
		if wxRsp.Code != wechatv3.Success {
			return nil, apperror.New(apperror.CodeWeChatAPICallFailed, wxRsp.Error)
		}
		jsapi, err := p.client.V3().PaySignOfJSAPI(appID, wxRsp.Response.PrepayId)
		if err != nil {
			return nil, apperror.Wrap(err, apperror.CodeWeChatAPICallFailed)
		}
		result = &service.ProviderPaymentResult{
			ThirdPartyNo: "",
			Status:       entity.PaymentStatusPending.String(),
			PayParams: map[string]any{
				"appId":     jsapi.AppId,
				"timeStamp": jsapi.TimeStamp,
				"nonceStr":  jsapi.NonceStr,
				"package":   jsapi.Package,
				"signType":  jsapi.SignType,
				"paySign":   jsapi.PaySign,
			},
			RawResponse: map[string]any{"prepay_id": wxRsp.Response.PrepayId},
		}

	default:
		return nil, apperror.New(apperror.CodeInvalidPaymentMethod, fmt.Sprintf("unsupported wechat method: %s", req.Method))
	}

	return result, nil
}

func (p *provider) QueryPayment(ctx context.Context, thirdPartyNo string) (*service.ProviderPaymentResult, error) {
	if !p.client.IsAvailable() {
		return nil, apperror.New(apperror.CodeWeChatAPICallFailed, "wechat client not available")
	}

	wxRsp, err := p.client.V3().V3TransactionQueryOrder(ctx, wechatv3.OutTradeNo, thirdPartyNo)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeWeChatAPICallFailed)
	}
	if wxRsp.Code != wechatv3.Success {
		return nil, apperror.New(apperror.CodeWeChatAPICallFailed, wxRsp.Error)
	}

	return &service.ProviderPaymentResult{
		ThirdPartyNo: wxRsp.Response.TransactionId,
		Status:       wxRsp.Response.TradeState,
		RawResponse: map[string]any{
			"trade_state":      wxRsp.Response.TradeState,
			"trade_state_desc": wxRsp.Response.TradeStateDesc,
		},
	}, nil
}

func (p *provider) Refund(ctx context.Context, req service.ProviderRefundRequest) (*service.ProviderRefundResult, error) {
	if !p.client.IsAvailable() {
		return nil, apperror.New(apperror.CodeWeChatAPICallFailed, "wechat client not available")
	}

	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", req.PaymentNo).
		Set("out_refund_no", req.RefundNo).
		Set("reason", req.Reason).
		Set("notify_url", req.NotifyURL).
		SetBodyMap("amount", func(bm gopay.BodyMap) {
			bm.Set("refund", req.Amount).
				Set("total", req.Amount).
				Set("currency", "CNY")
		})

	wxRsp, err := p.client.V3().V3Refund(ctx, bm)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeWeChatAPICallFailed)
	}
	if wxRsp.Code != wechatv3.Success {
		return nil, apperror.New(apperror.CodeWeChatAPICallFailed, wxRsp.Error)
	}

	return &service.ProviderRefundResult{
		ThirdPartyNo: wxRsp.Response.RefundId,
		Status:       wxRsp.Response.Status,
		RawResponse: map[string]any{
			"status": wxRsp.Response.Status,
		},
	}, nil
}

func (p *provider) QueryRefund(ctx context.Context, thirdPartyNo string) (*service.ProviderRefundResult, error) {
	if !p.client.IsAvailable() {
		return nil, apperror.New(apperror.CodeWeChatAPICallFailed, "wechat client not available")
	}

	bm := make(gopay.BodyMap)
	wxRsp, err := p.client.V3().V3RefundQuery(ctx, thirdPartyNo, bm)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeWeChatAPICallFailed)
	}
	if wxRsp.Code != wechatv3.Success {
		return nil, apperror.New(apperror.CodeWeChatAPICallFailed, wxRsp.Error)
	}

	return &service.ProviderRefundResult{
		ThirdPartyNo: wxRsp.Response.RefundId,
		Status:       wxRsp.Response.Status,
		RawResponse: map[string]any{
			"status": wxRsp.Response.Status,
		},
	}, nil
}

func (p *provider) VerifyNotify(ctx context.Context, body []byte, headers map[string]string) (map[string]string, error) {
	if !p.client.IsAvailable() {
		return nil, apperror.New(apperror.CodeWeChatAPICallFailed, "wechat client not available")
	}

	// Basic anti-replay: verify timestamp is within 5 minutes
	timestamp := headers["Wechatpay-Timestamp"]
	if timestamp != "" {
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err == nil {
			now := time.Now().Unix()
			if now-ts > 300 || ts-now > 300 {
				return nil, apperror.New(apperror.CodeWebhookInvalidSignature, "wechat notify timestamp expired")
			}
		}
	}

	var notifyReq wechatv3.V3NotifyReq
	if err := json.Unmarshal(body, &notifyReq); err != nil {
		return nil, apperror.Wrap(err, apperror.CodeWebhookInvalidSignature)
	}

	result, err := notifyReq.DecryptPayCipherText(p.client.cfg.APIV3Key)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.CodeWebhookInvalidSignature)
	}

	return map[string]string{
		"out_trade_no":   result.OutTradeNo,
		"transaction_id": result.TransactionId,
		"trade_state":    result.TradeState,
	}, nil
}
