package dto

type CreateRefundRequest struct {
	PaymentNo string `json:"payment_no" binding:"required"`
	Amount    int64  `json:"amount" binding:"required,gt=0"`
	Reason    string `json:"reason" binding:"required,max=256"`
}

type RefundResponse struct {
	RefundNo   string `json:"refund_no"`
	PaymentNo  string `json:"payment_no"`
	OrderNo    string `json:"order_no"`
	Channel    string `json:"channel"`
	Amount     int64  `json:"amount"`
	Reason     string `json:"reason"`
	Status     string `json:"status"`
	CreatedAt  int64  `json:"created_at"`
}
