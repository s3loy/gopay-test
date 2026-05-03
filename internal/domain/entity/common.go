package entity

import "fmt"

type OrderStatus int8

const (
	OrderStatusPending       OrderStatus = 0
	OrderStatusPaid          OrderStatus = 1
	OrderStatusPartialRefund OrderStatus = 2
	OrderStatusFullRefund    OrderStatus = 3
	OrderStatusClosed        OrderStatus = 4
	OrderStatusExpired       OrderStatus = 5
)

func (s OrderStatus) String() string {
	switch s {
	case OrderStatusPending:
		return "pending"
	case OrderStatusPaid:
		return "paid"
	case OrderStatusPartialRefund:
		return "partial_refund"
	case OrderStatusFullRefund:
		return "full_refund"
	case OrderStatusClosed:
		return "closed"
	case OrderStatusExpired:
		return "expired"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

type PaymentStatus int8

const (
	PaymentStatusPending    PaymentStatus = 0
	PaymentStatusSuccess    PaymentStatus = 1
	PaymentStatusFailed     PaymentStatus = 2
	PaymentStatusCancelled  PaymentStatus = 3
	PaymentStatusProcessing PaymentStatus = 4
)

func (s PaymentStatus) String() string {
	switch s {
	case PaymentStatusPending:
		return "pending"
	case PaymentStatusSuccess:
		return "success"
	case PaymentStatusFailed:
		return "failed"
	case PaymentStatusCancelled:
		return "cancelled"
	case PaymentStatusProcessing:
		return "processing"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

type RefundStatus int8

const (
	RefundStatusPending    RefundStatus = 0
	RefundStatusSuccess    RefundStatus = 1
	RefundStatusFailed     RefundStatus = 2
	RefundStatusProcessing RefundStatus = 3
	RefundStatusRejected   RefundStatus = 4
)

func (s RefundStatus) String() string {
	switch s {
	case RefundStatusPending:
		return "pending"
	case RefundStatusSuccess:
		return "success"
	case RefundStatusFailed:
		return "failed"
	case RefundStatusProcessing:
		return "processing"
	case RefundStatusRejected:
		return "rejected"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

type PaymentChannel string

const (
	ChannelWechat PaymentChannel = "wechat"
	ChannelAlipay PaymentChannel = "alipay"
)

type PaymentMethod string

const (
	MethodNative PaymentMethod = "native"
	MethodJSAPI  PaymentMethod = "jsapi"
	MethodPC     PaymentMethod = "pc"
	MethodWAP    PaymentMethod = "wap"
	MethodAPP    PaymentMethod = "app"
)
