package repository

import (
	"context"

	"github.com/s3loy/gopay/internal/domain/entity"
)

type RefundFilter struct {
	PaymentID uint64
	OrderID   uint64
	Status    entity.RefundStatus
	Page      int
	Size      int
}

type RefundRepository interface {
	Create(ctx context.Context, refund *entity.Refund) error
	GetByID(ctx context.Context, id uint64) (*entity.Refund, error)
	GetByRefundNo(ctx context.Context, refundNo string) (*entity.Refund, error)
	GetByPaymentID(ctx context.Context, paymentID uint64) ([]*entity.Refund, error)
	Update(ctx context.Context, refund *entity.Refund) error
	UpdateStatus(ctx context.Context, id uint64, status entity.RefundStatus) error
	GetTotalRefundAmount(ctx context.Context, paymentID uint64) (int64, error)
	List(ctx context.Context, filter RefundFilter) ([]*entity.Refund, int64, error)
}
