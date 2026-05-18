package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/s3loy/gopay/internal/domain/entity"
	"github.com/s3loy/gopay/internal/domain/repository"
	"github.com/s3loy/gopay/internal/domain/service"
	"github.com/s3loy/gopay/internal/pkg/apperror"
	"github.com/s3loy/gopay/internal/pkg/logger"
	"go.uber.org/zap"
)

type CreatePaymentRequest struct {
	OrderNo       string
	Channel       entity.PaymentChannel
	Method        entity.PaymentMethod
	ClientIP      string
	NotifyURL     string
	ReturnURL     string
	OpenID        string
	BuyerID       string
	ExpireMinutes int
}

type PaymentUsecase interface {
	Create(ctx context.Context, req CreatePaymentRequest) (*entity.Payment, map[string]any, error)
	Get(ctx context.Context, paymentNo string) (*entity.Payment, error)
	HandleWechatNotify(ctx context.Context, body []byte, headers map[string]string) error
	HandleAlipayNotify(ctx context.Context, params map[string]string) error
}

type paymentUsecase struct {
	orderRepo    repository.OrderRepository
	paymentRepo  repository.PaymentRepository
	providerFact service.PaymentProviderFactory
	txMgr        repository.TransactionManager
}

func NewPaymentUsecase(
	orderRepo repository.OrderRepository,
	paymentRepo repository.PaymentRepository,
	providerFact service.PaymentProviderFactory,
	txMgr repository.TransactionManager,
) PaymentUsecase {
	return &paymentUsecase{
		orderRepo:    orderRepo,
		paymentRepo:  paymentRepo,
		providerFact: providerFact,
		txMgr:        txMgr,
	}
}

func (u *paymentUsecase) Create(ctx context.Context, req CreatePaymentRequest) (*entity.Payment, map[string]any, error) {
	order, err := u.orderRepo.GetByOrderNo(ctx, req.OrderNo)
	if err != nil {
		return nil, nil, err
	}
	if !order.CanPay() {
		return nil, nil, apperror.New(apperror.CodeOrderCannotClose, "order cannot be paid")
	}

	provider, err := u.providerFact.Get(req.Channel)
	if err != nil {
		return nil, nil, err
	}

	if req.ExpireMinutes <= 0 {
		req.ExpireMinutes = 30
	}
	if req.ExpireMinutes > 1440 {
		req.ExpireMinutes = 1440
	}

	payment := &entity.Payment{
		PaymentNo: generatePaymentNo(),
		OrderID:   order.ID,
		OrderNo:   order.OrderNo,
		Channel:   req.Channel,
		Method:    req.Method,
		Amount:    order.Amount,
		Currency:  order.Currency,
		Status:    entity.PaymentStatusPending,
		ClientIP:  req.ClientIP,
		NotifyURL: req.NotifyURL,
		ReturnURL: req.ReturnURL,
		ExpireAt:  time.Now().Add(time.Duration(req.ExpireMinutes) * time.Minute),
	}

	if err := u.paymentRepo.Create(ctx, payment); err != nil {
		return nil, nil, err
	}

	result, err := provider.CreatePayment(ctx, service.ProviderPaymentRequest{
		OrderNo:       order.OrderNo,
		Subject:       order.Subject,
		Amount:        order.Amount,
		Currency:      order.Currency,
		Method:        req.Method,
		ClientIP:      req.ClientIP,
		NotifyURL:     req.NotifyURL,
		ReturnURL:     req.ReturnURL,
		OpenID:        req.OpenID,
		BuyerID:       req.BuyerID,
		ExpireMinutes: req.ExpireMinutes,
	})
	if err != nil {
		if upErr := u.paymentRepo.UpdateStatus(ctx, payment.ID, entity.PaymentStatusFailed); upErr != nil {
			logger.L().Error("update payment status to failed failed", zap.Error(upErr), zap.String("payment_no", payment.PaymentNo))
		}
		return nil, nil, err
	}

	if result.ThirdPartyNo != "" {
		payment.ThirdPartyNo = result.ThirdPartyNo
	}
	payment.ThirdPartyResp = result.RawResponse
	if err := u.paymentRepo.Update(ctx, payment); err != nil {
		logger.L().Warn("update payment after create failed", zap.Error(err), zap.String("payment_no", payment.PaymentNo))
	}

	logger.L().Info("payment created",
		zap.String("payment_no", payment.PaymentNo),
		zap.String("order_no", order.OrderNo),
		zap.String("channel", string(req.Channel)),
		zap.String("method", string(req.Method)),
	)

	return payment, result.PayParams, nil
}

func (u *paymentUsecase) Get(ctx context.Context, paymentNo string) (*entity.Payment, error) {
	return u.paymentRepo.GetByPaymentNo(ctx, paymentNo)
}

func (u *paymentUsecase) HandleWechatNotify(ctx context.Context, body []byte, headers map[string]string) error {
	provider, err := u.providerFact.Get(entity.ChannelWechat)
	if err != nil {
		return err
	}

	result, err := provider.VerifyNotify(ctx, body, headers)
	if err != nil {
		logger.L().Error("wechat notify verify failed", zap.Error(err))
		return err
	}

	outTradeNo := result["out_trade_no"]
	tradeState := result["trade_state"]
	transactionID := result["transaction_id"]

	logger.L().Info("wechat notify received",
		zap.String("out_trade_no", outTradeNo),
		zap.String("trade_state", tradeState),
		zap.String("transaction_id", transactionID),
	)

	return u.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		payment, err := u.paymentRepo.GetByPaymentNo(txCtx, outTradeNo)
		if err != nil {
			return err
		}

		if payment.Status == entity.PaymentStatusSuccess {
			logger.L().Warn("wechat notify duplicate", zap.String("payment_no", payment.PaymentNo))
			return nil
		}

		if tradeState == "SUCCESS" {
			payment.Status = entity.PaymentStatusSuccess
			payment.ThirdPartyNo = transactionID
			now := time.Now()
			payment.PaidAt = &now
			if err := u.paymentRepo.Update(txCtx, payment); err != nil {
				return err
			}

			order, err := u.orderRepo.GetByID(txCtx, payment.OrderID)
			if err != nil {
				return err
			}
			if err := order.MarkPaid(); err != nil {
				return err
			}
			if err := u.orderRepo.UpdateStatus(txCtx, order.ID, entity.OrderStatusPaid); err != nil {
				return err
			}

			logger.L().Info("payment success via wechat notify",
				zap.String("payment_no", payment.PaymentNo),
				zap.String("order_no", order.OrderNo),
				zap.String("transaction_id", transactionID),
			)
		}

		return nil
	})
}

func (u *paymentUsecase) HandleAlipayNotify(ctx context.Context, params map[string]string) error {
	provider, err := u.providerFact.Get(entity.ChannelAlipay)
	if err != nil {
		return err
	}

	result, err := provider.VerifyNotify(ctx, nil, params)
	if err != nil {
		logger.L().Error("alipay notify verify failed", zap.Error(err))
		return err
	}

	outTradeNo := result["out_trade_no"]
	tradeStatus := result["trade_status"]
	tradeNo := result["trade_no"]

	logger.L().Info("alipay notify received",
		zap.String("out_trade_no", outTradeNo),
		zap.String("trade_status", tradeStatus),
		zap.String("trade_no", tradeNo),
	)

	return u.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		payment, err := u.paymentRepo.GetByPaymentNo(txCtx, outTradeNo)
		if err != nil {
			return err
		}

		if payment.Status == entity.PaymentStatusSuccess {
			logger.L().Warn("alipay notify duplicate", zap.String("payment_no", payment.PaymentNo))
			return nil
		}

		if tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED" {
			payment.Status = entity.PaymentStatusSuccess
			payment.ThirdPartyNo = tradeNo
			now := time.Now()
			payment.PaidAt = &now
			if err := u.paymentRepo.Update(txCtx, payment); err != nil {
				return err
			}

			order, err := u.orderRepo.GetByID(txCtx, payment.OrderID)
			if err != nil {
				return err
			}
			if err := order.MarkPaid(); err != nil {
				return err
			}
			if err := u.orderRepo.UpdateStatus(txCtx, order.ID, entity.OrderStatusPaid); err != nil {
				return err
			}

			logger.L().Info("payment success via alipay notify",
				zap.String("payment_no", payment.PaymentNo),
				zap.String("order_no", order.OrderNo),
				zap.String("trade_no", tradeNo),
			)
		}

		return nil
	})
}

func generatePaymentNo() string {
	return fmt.Sprintf("PAY%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
}
