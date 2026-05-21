package entity

import "testing"

func TestOrderStatus_String(t *testing.T) {
	tests := []struct {
		status OrderStatus
		want   string
	}{
		{OrderStatusPending, "pending"},
		{OrderStatusPaid, "paid"},
		{OrderStatusPartialRefund, "partial_refund"},
		{OrderStatusFullRefund, "full_refund"},
		{OrderStatusClosed, "closed"},
		{OrderStatusExpired, "expired"},
		{OrderStatus(99), "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPaymentStatus_String(t *testing.T) {
	tests := []struct {
		status PaymentStatus
		want   string
	}{
		{PaymentStatusPending, "pending"},
		{PaymentStatusSuccess, "success"},
		{PaymentStatusFailed, "failed"},
		{PaymentStatusCancelled, "cancelled"},
		{PaymentStatusProcessing, "processing"},
		{PaymentStatusRefunded, "refunded"},
		{PaymentStatus(99), "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRefundStatus_String(t *testing.T) {
	tests := []struct {
		status RefundStatus
		want   string
	}{
		{RefundStatusPending, "pending"},
		{RefundStatusSuccess, "success"},
		{RefundStatusFailed, "failed"},
		{RefundStatusProcessing, "processing"},
		{RefundStatusRejected, "rejected"},
		{RefundStatus(99), "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
