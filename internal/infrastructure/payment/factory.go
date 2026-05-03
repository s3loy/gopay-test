package payment

import (
	"github.com/s3loy/gopay/internal/domain/entity"
	"github.com/s3loy/gopay/internal/domain/service"
	"github.com/s3loy/gopay/internal/infrastructure/payment/alipay"
	"github.com/s3loy/gopay/internal/infrastructure/payment/wechat"
	"github.com/s3loy/gopay/internal/pkg/apperror"
)

type factory struct {
	wechatProvider service.PaymentProvider
	alipayProvider service.PaymentProvider
}

func NewProviderFactory(wechatClient *wechat.Client, alipayClient *alipay.Client) service.PaymentProviderFactory {
	return &factory{
		wechatProvider: wechat.NewProvider(wechatClient),
		alipayProvider: alipay.NewProvider(alipayClient),
	}
}

func (f *factory) Get(channel entity.PaymentChannel) (service.PaymentProvider, error) {
	switch channel {
	case entity.ChannelWechat:
		return f.wechatProvider, nil
	case entity.ChannelAlipay:
		return f.alipayProvider, nil
	default:
		return nil, apperror.New(apperror.CodeInvalidPaymentChannel, "unsupported payment channel")
	}
}
