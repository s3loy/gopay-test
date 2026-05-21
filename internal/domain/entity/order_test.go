package entity

import (
	"testing"
	"time"

	"github.com/s3loy/gopay/internal/pkg/apperror"
)

func TestOrder_CanPay(t *testing.T) {
	tests := []struct {
		name     string
		status   OrderStatus
		expired  bool
		expected bool
	}{
		{"pending not expired", OrderStatusPending, false, true},
		{"pending expired", OrderStatusPending, true, false},
		{"paid", OrderStatusPaid, false, false},
		{"closed", OrderStatusClosed, false, false},
		{"expired status", OrderStatusExpired, false, false},
		{"partial refund", OrderStatusPartialRefund, false, false},
		{"full refund", OrderStatusFullRefund, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{Status: tt.status}
			if tt.expired {
				o.ExpiredAt = time.Now().Add(-time.Minute)
			} else {
				o.ExpiredAt = time.Now().Add(time.Minute)
			}
			if got := o.CanPay(); got != tt.expected {
				t.Errorf("CanPay() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOrder_CanClose(t *testing.T) {
	tests := []struct {
		name     string
		status   OrderStatus
		expected bool
	}{
		{"pending", OrderStatusPending, true},
		{"paid", OrderStatusPaid, false},
		{"closed", OrderStatusClosed, false},
		{"expired", OrderStatusExpired, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{Status: tt.status}
			if got := o.CanClose(); got != tt.expected {
				t.Errorf("CanClose() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOrder_CanRefund(t *testing.T) {
	tests := []struct {
		name     string
		status   OrderStatus
		expected bool
	}{
		{"paid", OrderStatusPaid, true},
		{"partial refund", OrderStatusPartialRefund, true},
		{"pending", OrderStatusPending, false},
		{"closed", OrderStatusClosed, false},
		{"full refund", OrderStatusFullRefund, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{Status: tt.status}
			if got := o.CanRefund(); got != tt.expected {
				t.Errorf("CanRefund() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOrder_MarkPaid(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		o := &Order{Status: OrderStatusPending, ExpiredAt: time.Now().Add(time.Minute)}
		if err := o.MarkPaid(); err != nil {
			t.Fatalf("MarkPaid() error = %v", err)
		}
		if o.Status != OrderStatusPaid {
			t.Errorf("Status = %v, want %v", o.Status, OrderStatusPaid)
		}
		if o.PaidAt == nil {
			t.Error("PaidAt should not be nil")
		}
	})

	t.Run("already paid", func(t *testing.T) {
		o := &Order{Status: OrderStatusPaid}
		err := o.MarkPaid()
		if err == nil {
			t.Fatal("MarkPaid() should return error")
		}
		if !apperror.Is(err, apperror.CodeOrderAlreadyPaid) {
			t.Errorf("error code = %v, want %v", err, apperror.CodeOrderAlreadyPaid)
		}
	})

	t.Run("expired", func(t *testing.T) {
		o := &Order{Status: OrderStatusPending, ExpiredAt: time.Now().Add(-time.Minute)}
		err := o.MarkPaid()
		if err == nil {
			t.Fatal("MarkPaid() should return error")
		}
	})
}

func TestOrder_MarkClosed(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		o := &Order{Status: OrderStatusPending}
		if err := o.MarkClosed(); err != nil {
			t.Fatalf("MarkClosed() error = %v", err)
		}
		if o.Status != OrderStatusClosed {
			t.Errorf("Status = %v, want %v", o.Status, OrderStatusClosed)
		}
	})

	t.Run("already closed", func(t *testing.T) {
		o := &Order{Status: OrderStatusClosed}
		err := o.MarkClosed()
		if err == nil {
			t.Fatal("MarkClosed() should return error")
		}
		if !apperror.Is(err, apperror.CodeOrderAlreadyClosed) {
			t.Errorf("error code = %v, want %v", err, apperror.CodeOrderAlreadyClosed)
		}
	})

	t.Run("already paid", func(t *testing.T) {
		o := &Order{Status: OrderStatusPaid}
		err := o.MarkClosed()
		if err == nil {
			t.Fatal("MarkClosed() should return error")
		}
	})
}

func TestOrder_IsExpired(t *testing.T) {
	t.Run("expired", func(t *testing.T) {
		o := &Order{ExpiredAt: time.Now().Add(-time.Minute)}
		if !o.IsExpired() {
			t.Error("IsExpired() = false, want true")
		}
	})

	t.Run("not expired", func(t *testing.T) {
		o := &Order{ExpiredAt: time.Now().Add(time.Minute)}
		if o.IsExpired() {
			t.Error("IsExpired() = true, want false")
		}
	})
}
