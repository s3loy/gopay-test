package repository

import (
	"context"

	"github.com/s3loy/gopay/internal/domain/entity"
)

type PaymentFilter struct {
	OrderID uint64
	Channel entity.PaymentChannel
	Status  entity.PaymentStatus
	Page    int
	Size    int
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *entity.Payment) error
	GetByID(ctx context.Context, id uint64) (*entity.Payment, error)
	GetByPaymentNo(ctx context.Context, paymentNo string) (*entity.Payment, error)
	GetByThirdPartyNo(ctx context.Context, thirdPartyNo string) (*entity.Payment, error)
	GetByOrderID(ctx context.Context, orderID uint64) ([]*entity.Payment, error)
	Update(ctx context.Context, payment *entity.Payment) error
	UpdateStatus(ctx context.Context, id uint64, status entity.PaymentStatus) error
	List(ctx context.Context, filter PaymentFilter) ([]*entity.Payment, int64, error)
}
