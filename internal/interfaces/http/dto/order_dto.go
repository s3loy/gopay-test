package dto

type CreateOrderRequest struct {
	UserID        uint64 `json:"user_id" binding:"required,gt=0"`
	Subject       string `json:"subject" binding:"required,max=256"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"omitempty,oneof=CNY"`
	Description   string `json:"description" binding:"omitempty,max=512"`
	ExpireMinutes int    `json:"expire_minutes" binding:"omitempty,min=1,max=1440"`
}

type OrderResponse struct {
	OrderNo     string                 `json:"order_no"`
	UserID      uint64                 `json:"user_id"`
	Subject     string                 `json:"subject"`
	Amount      int64                  `json:"amount"`
	Currency    string                 `json:"currency"`
	Status      string                 `json:"status"`
	ExpiredAt   int64                  `json:"expired_at"`
	PaidAt      *int64                 `json:"paid_at,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   int64                  `json:"created_at"`
}

type CloseOrderRequest struct {
	OrderNo string `json:"order_no" uri:"order_no" binding:"required"`
}
