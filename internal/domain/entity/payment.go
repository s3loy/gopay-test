package entity

import "time"

type Payment struct {
	ID             uint64
	PaymentNo      string
	OrderID        uint64
	OrderNo        string
	Channel        PaymentChannel
	Method         PaymentMethod
	Amount         int64 // 分
	Currency       string
	Status         PaymentStatus
	ThirdPartyNo   string
	ThirdPartyResp map[string]any
	ClientIP       string
	NotifyURL      string
	ReturnURL      string
	ExpireAt       time.Time
	PaidAt         *time.Time
	NotifyAt       *time.Time
	NotifyCount    int16
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (p *Payment) IsSuccess() bool {
	return p.Status == PaymentStatusSuccess
}

func (p *Payment) CanRefund() bool {
	return p.Status == PaymentStatusSuccess
}

func (p *Payment) IsExpired() bool {
	return time.Now().After(p.ExpireAt)
}
