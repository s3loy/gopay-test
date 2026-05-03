package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/s3loy/gopay/internal/domain/entity"
	"github.com/s3loy/gopay/internal/interfaces/http/dto"
	"github.com/s3loy/gopay/internal/pkg/apperror"
	"github.com/s3loy/gopay/internal/pkg/response"
	"github.com/s3loy/gopay/internal/pkg/validator"
	"github.com/s3loy/gopay/internal/usecase"
)

type PaymentHandler struct {
	paymentUC usecase.PaymentUsecase
}

func NewPaymentHandler(paymentUC usecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{paymentUC: paymentUC}
}

func (h *PaymentHandler) Create(c *gin.Context) {
	var req dto.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidParams, err.Error()))
		return
	}
	if err := validator.ValidateStruct(&req); err != nil {
		response.Error(c, err)
		return
	}

	payment, payParams, err := h.paymentUC.Create(c.Request.Context(), usecase.CreatePaymentRequest{
		OrderNo:       req.OrderNo,
		Channel:       entity.PaymentChannel(req.Channel),
		Method:        entity.PaymentMethod(req.Method),
		ClientIP:      req.ClientIP,
		NotifyURL:     req.NotifyURL,
		ReturnURL:     req.ReturnURL,
		OpenID:        req.OpenID,
		BuyerID:       req.BuyerID,
		ExpireMinutes: req.ExpireMinutes,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	resp := toPaymentResponse(payment)
	resp.PayParams = payParams
	response.OK(c, resp)
}

func (h *PaymentHandler) Get(c *gin.Context) {
	paymentNo := c.Param("payment_no")
	if paymentNo == "" {
		response.ErrorWithCode(c, apperror.CodeMissingParam)
		return
	}

	payment, err := h.paymentUC.Get(c.Request.Context(), paymentNo)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, toPaymentResponse(payment))
}

func toPaymentResponse(payment *entity.Payment) dto.PaymentResponse {
	return dto.PaymentResponse{
		PaymentNo: payment.PaymentNo,
		OrderNo:   payment.OrderNo,
		Channel:   string(payment.Channel),
		Method:    string(payment.Method),
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		Status:    payment.Status.String(),
		ExpireAt:  payment.ExpireAt.Unix(),
		CreatedAt: payment.CreatedAt.Unix(),
	}
}
