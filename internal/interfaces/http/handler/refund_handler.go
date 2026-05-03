package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/s3loy/gopay/internal/interfaces/http/dto"
	"github.com/s3loy/gopay/internal/pkg/apperror"
	"github.com/s3loy/gopay/internal/pkg/response"
	"github.com/s3loy/gopay/internal/pkg/validator"
	"github.com/s3loy/gopay/internal/usecase"
)

type RefundHandler struct {
	refundUC usecase.RefundUsecase
}

func NewRefundHandler(refundUC usecase.RefundUsecase) *RefundHandler {
	return &RefundHandler{refundUC: refundUC}
}

func (h *RefundHandler) Create(c *gin.Context) {
	var req dto.CreateRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidParams, err.Error()))
		return
	}
	if err := validator.ValidateStruct(&req); err != nil {
		response.Error(c, err)
		return
	}

	refund, err := h.refundUC.Create(c.Request.Context(), usecase.CreateRefundRequest{
		PaymentNo: req.PaymentNo,
		Amount:    req.Amount,
		Reason:    req.Reason,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, dto.RefundResponse{
		RefundNo:  refund.RefundNo,
		PaymentNo: refund.PaymentNo,
		OrderNo:   refund.OrderNo,
		Channel:   string(refund.Channel),
		Amount:    refund.Amount,
		Reason:    refund.Reason,
		Status:    refund.Status.String(),
		CreatedAt: refund.CreatedAt.Unix(),
	})
}

func (h *RefundHandler) Get(c *gin.Context) {
	refundNo := c.Param("refund_no")
	if refundNo == "" {
		response.ErrorWithCode(c, apperror.CodeMissingParam)
		return
	}

	refund, err := h.refundUC.Get(c.Request.Context(), refundNo)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, dto.RefundResponse{
		RefundNo:  refund.RefundNo,
		PaymentNo: refund.PaymentNo,
		OrderNo:   refund.OrderNo,
		Channel:   string(refund.Channel),
		Amount:    refund.Amount,
		Reason:    refund.Reason,
		Status:    refund.Status.String(),
		CreatedAt: refund.CreatedAt.Unix(),
	})
}
