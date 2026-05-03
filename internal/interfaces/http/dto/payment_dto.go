package dto

type CreatePaymentRequest struct {
	OrderNo       string `json:"order_no" binding:"required"`
	Channel       string `json:"channel" binding:"required,oneof=wechat alipay"`
	Method        string `json:"method" binding:"required,oneof=native jsapi pc wap app"`
	ClientIP      string `json:"client_ip" binding:"required,ip"`
	NotifyURL     string `json:"notify_url" binding:"omitempty,url"`
	ReturnURL     string `json:"return_url" binding:"omitempty,url"`
	OpenID        string `json:"openid,omitempty"`
	BuyerID       string `json:"buyer_id,omitempty"`
	ExpireMinutes int    `json:"expire_minutes" binding:"omitempty,min=1,max=1440"`
}

type PaymentResponse struct {
	PaymentNo string                 `json:"payment_no"`
	OrderNo   string                 `json:"order_no"`
	Channel   string                 `json:"channel"`
	Method    string                 `json:"method"`
	Amount    int64                  `json:"amount"`
	Currency  string                 `json:"currency"`
	Status    string                 `json:"status"`
	PayParams map[string]interface{} `json:"pay_params,omitempty"`
	ExpireAt  int64                  `json:"expire_at"`
	CreatedAt int64                  `json:"created_at"`
}

type PaymentStatusRequest struct {
	PaymentNo string `uri:"payment_no" binding:"required"`
}
