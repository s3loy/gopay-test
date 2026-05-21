package postgresql

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value any) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

type OrderModel struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"`
	OrderNo     string    `gorm:"uniqueIndex;size:32;not null"`
	UserID      uint64    `gorm:"index:idx_orders_user_status;not null"`
	Subject     string    `gorm:"size:256;not null"`
	Amount      int64     `gorm:"not null"`
	Currency    string    `gorm:"size:3;default:'CNY';not null"`
	Status      int8      `gorm:"index:idx_orders_status_created;not null"`
	ExpiredAt   time.Time `gorm:"not null"`
	PaidAt      *time.Time
	Description string  `gorm:"size:512"`
	Metadata    JSONMap `gorm:"type:jsonb;default:'{}'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (OrderModel) TableName() string {
	return "orders"
}

type PaymentModel struct {
	ID             uint64  `gorm:"primaryKey;autoIncrement"`
	PaymentNo      string  `gorm:"uniqueIndex;size:32;not null"`
	OrderID        uint64  `gorm:"index:idx_payments_order_id;not null"`
	OrderNo        string  `gorm:"size:32;not null"`
	Channel        string  `gorm:"index:idx_payments_status_channel;size:16;not null"`
	Method         string  `gorm:"size:16;not null"`
	Amount         int64   `gorm:"not null"`
	Currency       string  `gorm:"size:3;default:'CNY';not null"`
	Status         int8    `gorm:"index:idx_payments_status_created;not null"`
	ThirdPartyNo   string  `gorm:"index:idx_payments_third_party;size:64"`
	ThirdPartyResp JSONMap `gorm:"type:jsonb;default:'{}'"`
	ClientIP       string  `gorm:"size:64"`
	NotifyURL      string
	ReturnURL      string
	ExpireAt       time.Time `gorm:"not null"`
	PaidAt         *time.Time
	NotifyAt       *time.Time
	NotifyCount    int16 `gorm:"default:0"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (PaymentModel) TableName() string {
	return "payments"
}

type RefundModel struct {
	ID             uint64  `gorm:"primaryKey;autoIncrement"`
	RefundNo       string  `gorm:"uniqueIndex;size:32;not null"`
	PaymentID      uint64  `gorm:"index:idx_refunds_payment_id;not null"`
	PaymentNo      string  `gorm:"size:32;not null"`
	OrderID        uint64  `gorm:"index:idx_refunds_order_id;not null"`
	OrderNo        string  `gorm:"size:32;not null"`
	Channel        string  `gorm:"index:idx_refunds_status_channel;size:16;not null"`
	Amount         int64   `gorm:"not null"`
	Reason         string  `gorm:"size:256;not null"`
	Status         int8    `gorm:"not null"`
	ThirdPartyNo   string  `gorm:"size:64"`
	ThirdPartyResp JSONMap `gorm:"type:jsonb;default:'{}'"`
	NotifyAt       *time.Time
	NotifyCount    int16 `gorm:"default:0"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (RefundModel) TableName() string {
	return "refunds"
}
