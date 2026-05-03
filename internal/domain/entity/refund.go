package entity

import "time"

type Refund struct {
	ID             uint64
	RefundNo       string
	PaymentID      uint64
	PaymentNo      string
	OrderID        uint64
	OrderNo        string
	Channel        PaymentChannel
	Amount         int64 // 分
	Reason         string
	Status         RefundStatus
	ThirdPartyNo   string
	ThirdPartyResp map[string]any
	NotifyAt       *time.Time
	NotifyCount    int16
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
