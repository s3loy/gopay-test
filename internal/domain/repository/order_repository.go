package repository

import (
	"context"

	"github.com/s3loy/gopay/internal/domain/entity"
)

type OrderFilter struct {
	UserID uint64
	Status entity.OrderStatus
	Page   int
	Size   int
}

type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	GetByID(ctx context.Context, id uint64) (*entity.Order, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*entity.Order, error)
	Update(ctx context.Context, order *entity.Order) error
	UpdateStatus(ctx context.Context, id uint64, status entity.OrderStatus) error
	List(ctx context.Context, filter OrderFilter) ([]*entity.Order, int64, error)
}
