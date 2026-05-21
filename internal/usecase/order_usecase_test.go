package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/s3loy/gopay/internal/domain/entity"
	"github.com/s3loy/gopay/internal/domain/repository"
	"github.com/s3loy/gopay/internal/pkg/apperror"
)

type mockOrderRepo struct {
	order   *entity.Order
	err     error
	created *entity.Order
	updated bool
}

func (m *mockOrderRepo) Create(ctx context.Context, order *entity.Order) error {
	m.created = order
	if m.err != nil {
		return m.err
	}
	order.ID = 1
	return nil
}

func (m *mockOrderRepo) GetByID(ctx context.Context, id uint64) (*entity.Order, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.order, nil
}

func (m *mockOrderRepo) GetByOrderNo(ctx context.Context, orderNo string) (*entity.Order, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.order, nil
}

func (m *mockOrderRepo) Update(ctx context.Context, order *entity.Order) error {
	m.updated = true
	return m.err
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, id uint64, status entity.OrderStatus) error {
	m.updated = true
	return m.err
}

func (m *mockOrderRepo) List(ctx context.Context, filter repository.OrderFilter) ([]*entity.Order, int64, error) {
	return nil, 0, m.err
}

func TestOrderUsecase_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo)
		order, err := uc.Create(context.Background(), 1, "test", 100, "CNY", "desc", 30)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if order.Amount != 100 {
			t.Errorf("Amount = %d, want 100", order.Amount)
		}
		if order.Currency != "CNY" {
			t.Errorf("Currency = %s, want CNY", order.Currency)
		}
		if order.Status != entity.OrderStatusPending {
			t.Errorf("Status = %v, want pending", order.Status)
		}
	})

	t.Run("amount must be positive", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo)
		_, err := uc.Create(context.Background(), 1, "test", 0, "CNY", "desc", 30)
		if err == nil {
			t.Fatal("expected error for zero amount")
		}
		if !apperror.Is(err, apperror.CodeOrderAmountInvalid) {
			t.Errorf("error code mismatch: %v", err)
		}
	})

	t.Run("default currency", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo)
		order, err := uc.Create(context.Background(), 1, "test", 100, "", "desc", 30)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if order.Currency != "CNY" {
			t.Errorf("Currency = %s, want CNY", order.Currency)
		}
	})

	t.Run("default expire minutes", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo)
		order, err := uc.Create(context.Background(), 1, "test", 100, "CNY", "desc", 0)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if time.Until(order.ExpiredAt) < 25*time.Minute {
			t.Error("ExpiredAt should be ~30 minutes from now")
		}
	})

	t.Run("expire minutes capped at 1440", func(t *testing.T) {
		repo := &mockOrderRepo{}
		uc := NewOrderUsecase(repo)
		order, err := uc.Create(context.Background(), 1, "test", 100, "CNY", "desc", 9999)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if time.Until(order.ExpiredAt) > 1441*time.Minute {
			t.Error("ExpiredAt should be capped at 1440 minutes")
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockOrderRepo{err: errors.New("db error")}
		uc := NewOrderUsecase(repo)
		_, err := uc.Create(context.Background(), 1, "test", 100, "CNY", "desc", 30)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestOrderUsecase_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := &entity.Order{OrderNo: "ORD123", Status: entity.OrderStatusPending}
		repo := &mockOrderRepo{order: expected}
		uc := NewOrderUsecase(repo)
		order, err := uc.Get(context.Background(), "ORD123")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if order.OrderNo != "ORD123" {
			t.Errorf("OrderNo = %s, want ORD123", order.OrderNo)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &mockOrderRepo{err: apperror.New(apperror.CodeOrderNotFound, "not found")}
		uc := NewOrderUsecase(repo)
		_, err := uc.Get(context.Background(), "ORD999")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestOrderUsecase_Close(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockOrderRepo{order: &entity.Order{ID: 1, OrderNo: "ORD123", Status: entity.OrderStatusPending}}
		uc := NewOrderUsecase(repo)
		if err := uc.Close(context.Background(), "ORD123"); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		if !repo.updated {
			t.Error("UpdateStatus should be called")
		}
	})

	t.Run("already paid cannot close", func(t *testing.T) {
		repo := &mockOrderRepo{order: &entity.Order{ID: 1, OrderNo: "ORD123", Status: entity.OrderStatusPaid}}
		uc := NewOrderUsecase(repo)
		err := uc.Close(context.Background(), "ORD123")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("order not found", func(t *testing.T) {
		repo := &mockOrderRepo{err: apperror.New(apperror.CodeOrderNotFound, "not found")}
		uc := NewOrderUsecase(repo)
		err := uc.Close(context.Background(), "ORD999")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("update status error", func(t *testing.T) {
		repo := &mockOrderRepo{order: &entity.Order{ID: 1, OrderNo: "ORD123", Status: entity.OrderStatusPending}, err: errors.New("db error")}
		uc := NewOrderUsecase(repo)
		err := uc.Close(context.Background(), "ORD123")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
