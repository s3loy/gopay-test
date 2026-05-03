package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/s3loy/gopay/internal/domain/entity"
	"github.com/s3loy/gopay/internal/domain/repository"
	"github.com/s3loy/gopay/internal/pkg/apperror"
)

type OrderUsecase interface {
	Create(ctx context.Context, userID uint64, subject string, amount int64, currency string, description string, expireMinutes int) (*entity.Order, error)
	Get(ctx context.Context, orderNo string) (*entity.Order, error)
	Close(ctx context.Context, orderNo string) error
}

type orderUsecase struct {
	orderRepo repository.OrderRepository
}

func NewOrderUsecase(orderRepo repository.OrderRepository) OrderUsecase {
	return &orderUsecase{orderRepo: orderRepo}
}

func (u *orderUsecase) Create(ctx context.Context, userID uint64, subject string, amount int64, currency string, description string, expireMinutes int) (*entity.Order, error) {
	if amount <= 0 {
		return nil, apperror.New(apperror.CodeOrderAmountInvalid, "amount must be greater than 0")
	}
	if currency == "" {
		currency = "CNY"
	}
	if expireMinutes <= 0 {
		expireMinutes = 30
	}

	order := &entity.Order{
		OrderNo:     generateOrderNo(),
		UserID:      userID,
		Subject:     subject,
		Amount:      amount,
		Currency:    currency,
		Status:      entity.OrderStatusPending,
		ExpiredAt:   time.Now().Add(time.Duration(expireMinutes) * time.Minute),
		Description: description,
		Metadata:    make(map[string]any),
	}

	if err := u.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (u *orderUsecase) Get(ctx context.Context, orderNo string) (*entity.Order, error) {
	return u.orderRepo.GetByOrderNo(ctx, orderNo)
}

func (u *orderUsecase) Close(ctx context.Context, orderNo string) error {
	order, err := u.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return err
	}
	if err := order.MarkClosed(); err != nil {
		return err
	}
	return u.orderRepo.UpdateStatus(ctx, order.ID, entity.OrderStatusClosed)
}

func generateOrderNo() string {
	return fmt.Sprintf("ORD%s%d", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
}
