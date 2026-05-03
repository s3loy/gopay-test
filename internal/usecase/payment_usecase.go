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
}

func NewPaymentUsecase(
	orderRepo repository.OrderRepository,
	paymentRepo repository.PaymentRepository,
	providerFact service.PaymentProviderFactory,
) PaymentUsecase {
	return &paymentUsecase{
		orderRepo:    orderRepo,
		paymentRepo:  paymentRepo,
		providerFact: providerFact,
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
		_ = u.paymentRepo.UpdateStatus(ctx, payment.ID, entity.PaymentStatusFailed)
		return nil, nil, err
	}

	if result.ThirdPartyNo != "" {
		payment.ThirdPartyNo = result.ThirdPartyNo
	}
	payment.ThirdPartyResp = result.RawResponse
	if err := u.paymentRepo.Update(ctx, payment); err != nil {
		logger.L().Warn("update payment after create failed", zap.Error(err))
	}

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
		return err
	}

	outTradeNo := result["out_trade_no"]
	tradeState := result["trade_state"]
	transactionID := result["transaction_id"]

	payment, err := u.paymentRepo.GetByPaymentNo(ctx, outTradeNo)
	if err != nil {
		return err
	}

	if payment.Status == entity.PaymentStatusSuccess {
		return nil
	}

	if tradeState == "SUCCESS" {
		payment.Status = entity.PaymentStatusSuccess
		payment.ThirdPartyNo = transactionID
		now := time.Now()
		payment.PaidAt = &now
		if err := u.paymentRepo.Update(ctx, payment); err != nil {
			return err
		}

		order, _ := u.orderRepo.GetByID(ctx, payment.OrderID)
		if order != nil {
			_ = order.MarkPaid()
			_ = u.orderRepo.UpdateStatus(ctx, order.ID, entity.OrderStatusPaid)
		}
	}

	return nil
}

func (u *paymentUsecase) HandleAlipayNotify(ctx context.Context, params map[string]string) error {
	provider, err := u.providerFact.Get(entity.ChannelAlipay)
	if err != nil {
		return err
	}

	result, err := provider.VerifyNotify(ctx, nil, params)
	if err != nil {
		return err
	}

	outTradeNo := result["out_trade_no"]
	tradeStatus := result["trade_status"]
	tradeNo := result["trade_no"]

	payment, err := u.paymentRepo.GetByPaymentNo(ctx, outTradeNo)
	if err != nil {
		return err
	}

	if payment.Status == entity.PaymentStatusSuccess {
		return nil
	}

	if tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED" {
		payment.Status = entity.PaymentStatusSuccess
		payment.ThirdPartyNo = tradeNo
		now := time.Now()
		payment.PaidAt = &now
		if err := u.paymentRepo.Update(ctx, payment); err != nil {
			return err
		}

		order, _ := u.orderRepo.GetByID(ctx, payment.OrderID)
		if order != nil {
			_ = order.MarkPaid()
			_ = u.orderRepo.UpdateStatus(ctx, order.ID, entity.OrderStatusPaid)
		}
	}

	return nil
}

func generatePaymentNo() string {
	return fmt.Sprintf("PAY%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
}
