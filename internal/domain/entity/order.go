package entity

import (
	"time"

	"github.com/s3loy/gopay/internal/pkg/apperror"
)

type Order struct {
	ID          uint64
	OrderNo     string
	UserID      uint64
	Subject     string
	Amount      int64 // 分
	Currency    string
	Status      OrderStatus
	ExpiredAt   time.Time
	PaidAt      *time.Time
	Description string
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (o *Order) IsExpired() bool {
	return time.Now().After(o.ExpiredAt)
}

func (o *Order) CanPay() bool {
	return o.Status == OrderStatusPending && !o.IsExpired()
}

func (o *Order) CanClose() bool {
	return o.Status == OrderStatusPending
}

func (o *Order) CanRefund() bool {
	return o.Status == OrderStatusPaid || o.Status == OrderStatusPartialRefund
}

func (o *Order) MarkPaid() error {
	if !o.CanPay() {
		if o.Status == OrderStatusPaid {
			return apperror.New(apperror.CodeOrderAlreadyPaid, "order already paid")
		}
		return apperror.New(apperror.CodeOrderCannotClose, "order cannot be paid")
	}
	o.Status = OrderStatusPaid
	now := time.Now()
	o.PaidAt = &now
	return nil
}

func (o *Order) MarkClosed() error {
	if !o.CanClose() {
		return apperror.New(apperror.CodeOrderAlreadyClosed, "order already closed")
	}
	o.Status = OrderStatusClosed
	return nil
}
