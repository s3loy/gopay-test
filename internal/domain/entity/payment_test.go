package entity

import (
	"testing"
	"time"
)

func TestPayment_IsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		status   PaymentStatus
		expected bool
	}{
		{"success", PaymentStatusSuccess, true},
		{"pending", PaymentStatusPending, false},
		{"failed", PaymentStatusFailed, false},
		{"cancelled", PaymentStatusCancelled, false},
		{"processing", PaymentStatusProcessing, false},
		{"refunded", PaymentStatusRefunded, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Payment{Status: tt.status}
			if got := p.IsSuccess(); got != tt.expected {
				t.Errorf("IsSuccess() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPayment_CanRefund(t *testing.T) {
	tests := []struct {
		name     string
		status   PaymentStatus
		expected bool
	}{
		{"success", PaymentStatusSuccess, true},
		{"pending", PaymentStatusPending, false},
		{"failed", PaymentStatusFailed, false},
		{"refunded", PaymentStatusRefunded, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Payment{Status: tt.status}
			if got := p.CanRefund(); got != tt.expected {
				t.Errorf("CanRefund() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPayment_IsExpired(t *testing.T) {
	t.Run("expired", func(t *testing.T) {
		p := &Payment{ExpireAt: time.Now().Add(-time.Minute)}
		if !p.IsExpired() {
			t.Error("IsExpired() = false, want true")
		}
	})

	t.Run("not expired", func(t *testing.T) {
		p := &Payment{ExpireAt: time.Now().Add(time.Minute)}
		if p.IsExpired() {
			t.Error("IsExpired() = true, want false")
		}
	})
}
