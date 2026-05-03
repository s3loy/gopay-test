package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/s3loy/gopay/internal/domain/entity"
	"github.com/s3loy/gopay/internal/domain/repository"
	"github.com/s3loy/gopay/internal/domain/service"
	"github.com/s3loy/gopay/internal/pkg/apperror"
)

type CreateRefundRequest struct {
	PaymentNo string
	Amount    int64
	Reason    string
}

type RefundUsecase interface {
	Create(ctx context.Context, req CreateRefundRequest) (*entity.Refund, error)
	Get(ctx context.Context, refundNo string) (*entity.Refund, error)
}

type refundUsecase struct {
	paymentRepo  repository.PaymentRepository
	refundRepo   repository.RefundRepository
	providerFact service.PaymentProviderFactory
}

func NewRefundUsecase(
	paymentRepo repository.PaymentRepository,
	refundRepo repository.RefundRepository,
	providerFact service.PaymentProviderFactory,
) RefundUsecase {
	return &refundUsecase{
		paymentRepo:  paymentRepo,
		refundRepo:   refundRepo,
		providerFact: providerFact,
	}
}

func (u *refundUsecase) Create(ctx context.Context, req CreateRefundRequest) (*entity.Refund, error) {
	payment, err := u.paymentRepo.GetByPaymentNo(ctx, req.PaymentNo)
	if err != nil {
		return nil, err
	}
	if !payment.CanRefund() {
		return nil, apperror.New(apperror.CodePaymentNotRefundable, "payment not refundable")
	}

	if req.Amount <= 0 {
		return nil, apperror.New(apperror.CodeRefundAmountExceedsPayment, "refund amount must be greater than 0")
	}
	if req.Amount > payment.Amount {
		return nil, apperror.New(apperror.CodeRefundAmountExceedsPayment, "refund amount exceeds payment amount")
	}

	totalRefunded, err := u.refundRepo.GetTotalRefundAmount(ctx, payment.ID)
	if err != nil {
		return nil, err
	}

	remaining := payment.Amount - totalRefunded
	if req.Amount > remaining {
		return nil, apperror.New(apperror.CodeRefundAmountExceedsRemaining, "refund amount exceeds remaining refundable amount")
	}

	provider, err := u.providerFact.Get(payment.Channel)
	if err != nil {
		return nil, err
	}

	refund := &entity.Refund{
		RefundNo:  generateRefundNo(),
		PaymentID: payment.ID,
		PaymentNo: payment.PaymentNo,
		OrderID:   payment.OrderID,
		OrderNo:   payment.OrderNo,
		Channel:   payment.Channel,
		Amount:    req.Amount,
		Reason:    req.Reason,
		Status:    entity.RefundStatusPending,
	}

	if err := u.refundRepo.Create(ctx, refund); err != nil {
		return nil, err
	}

	result, err := provider.Refund(ctx, service.ProviderRefundRequest{
		PaymentNo:    payment.PaymentNo,
		ThirdPartyNo: payment.ThirdPartyNo,
		RefundNo:     refund.RefundNo,
		Amount:       req.Amount,
		Reason:       req.Reason,
		NotifyURL:    "",
	})
	if err != nil {
		_ = u.refundRepo.UpdateStatus(ctx, refund.ID, entity.RefundStatusFailed)
		return nil, err
	}

	refund.ThirdPartyNo = result.ThirdPartyNo
	refund.ThirdPartyResp = result.RawResponse
	refund.Status = entity.RefundStatusSuccess
	if err := u.refundRepo.Update(ctx, refund); err != nil {
		return nil, err
	}

	// Update order status if fully refunded
	totalRefunded, _ = u.refundRepo.GetTotalRefundAmount(ctx, payment.ID)
	if totalRefunded >= payment.Amount {
		_ = u.paymentRepo.UpdateStatus(ctx, payment.ID, entity.PaymentStatusFailed)
	}

	return refund, nil
}

func (u *refundUsecase) Get(ctx context.Context, refundNo string) (*entity.Refund, error) {
	return u.refundRepo.GetByRefundNo(ctx, refundNo)
}

func generateRefundNo() string {
	return fmt.Sprintf("REF%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
}
