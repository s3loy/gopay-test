//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/s3loy/gopay/internal/infrastructure/payment"
	"github.com/s3loy/gopay/internal/infrastructure/payment/alipay"
	"github.com/s3loy/gopay/internal/infrastructure/payment/wechat"
	"github.com/s3loy/gopay/internal/infrastructure/persistence/postgresql"
	"github.com/s3loy/gopay/internal/interfaces/http/handler"
	"github.com/s3loy/gopay/internal/interfaces/http/router"
	"github.com/s3loy/gopay/internal/pkg/config"
	"github.com/s3loy/gopay/internal/pkg/logger"
	"github.com/s3loy/gopay/internal/usecase"
)

func initializeApp(cfgPath string) (*app, error) {
	wire.Build(
		config.Load,
		logger.Init,
		postgresql.NewDB,
		wechat.NewClient,
		alipay.NewClient,
		payment.NewProviderFactory,
		postgresql.NewOrderRepository,
		postgresql.NewPaymentRepository,
		postgresql.NewRefundRepository,
		usecase.NewOrderUsecase,
		usecase.NewPaymentUsecase,
		usecase.NewRefundUsecase,
		handler.NewOrderHandler,
		handler.NewPaymentHandler,
		handler.NewRefundHandler,
		handler.NewWebhookHandler,
		handler.NewHealthHandler,
		router.NewRouter,
		newApp,
	)
	return nil, nil
}
