package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/s3loy/gopay/internal/infrastructure/payment"
	"github.com/s3loy/gopay/internal/infrastructure/payment/alipay"
	"github.com/s3loy/gopay/internal/infrastructure/payment/wechat"
	"github.com/s3loy/gopay/internal/infrastructure/persistence/postgresql"
	"github.com/s3loy/gopay/internal/interfaces/http/handler"
	"github.com/s3loy/gopay/internal/interfaces/http/router"
	"github.com/s3loy/gopay/internal/pkg/config"
	"github.com/s3loy/gopay/internal/pkg/logger"
	"github.com/s3loy/gopay/internal/usecase"
	"go.uber.org/zap"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "configs/config.dev.yaml", "config file path")
	flag.Parse()

	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config failed: %v\n", err)
		os.Exit(1)
	}

	// Init logger
	log, err := logger.Init(cfg.Log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger failed: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = log.Sync() //nolint:errcheck // sync error is safe to ignore during shutdown
	}()

	// Init database
	db, err := postgresql.NewDB(cfg.Database)
	if err != nil {
		logger.L().Fatal("init database failed", zap.Error(err))
	}

	// Init payment clients
	wechatClient, err := wechat.NewClient(cfg.Payment.Wechat)
	if err != nil {
		logger.L().Warn("init wechat client failed, wechat payment unavailable", zap.Error(err))
	}
	alipayClient, err := alipay.NewClient(cfg.Payment.Alipay)
	if err != nil {
		logger.L().Warn("init alipay client failed, alipay payment unavailable", zap.Error(err))
	}

	// Init provider factory
	providerFact := payment.NewProviderFactory(wechatClient, alipayClient)

	// Init repositories
	orderRepo := postgresql.NewOrderRepository(db)
	paymentRepo := postgresql.NewPaymentRepository(db)
	refundRepo := postgresql.NewRefundRepository(db)

	// Init transaction manager
	txMgr := postgresql.NewTransactionManager(db)

	// Init usecases
	orderUC := usecase.NewOrderUsecase(orderRepo)
	paymentUC := usecase.NewPaymentUsecase(orderRepo, paymentRepo, providerFact, txMgr)
	refundUC := usecase.NewRefundUsecase(paymentRepo, refundRepo, providerFact, txMgr)

	// Init handlers
	orderHandler := handler.NewOrderHandler(orderUC)
	paymentHandler := handler.NewPaymentHandler(paymentUC)
	refundHandler := handler.NewRefundHandler(refundUC)
	webhookHandler := handler.NewWebhookHandler(paymentUC)
	healthHandler := handler.NewHealthHandler()

	// Init router
	r := router.NewRouter(orderHandler, paymentHandler, refundHandler, webhookHandler, healthHandler, cfg.App.CORSOrigins)

	// Setup gin
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	r.Register(engine)

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.L().Info("server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L().Fatal("server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.L().Info("server shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.L().Error("server shutdown failed", zap.Error(err))
	}
	logger.L().Info("server stopped")
}
