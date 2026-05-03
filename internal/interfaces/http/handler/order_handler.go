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

type OrderHandler struct {
	orderUC usecase.OrderUsecase
}

func NewOrderHandler(orderUC usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{orderUC: orderUC}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperror.New(apperror.CodeInvalidParams, err.Error()))
		return
	}
	if err := validator.ValidateStruct(&req); err != nil {
		response.Error(c, err)
		return
	}

	order, err := h.orderUC.Create(c.Request.Context(), req.UserID, req.Subject, req.Amount, req.Currency, req.Description, req.ExpireMinutes)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, toOrderResponse(order))
}

func (h *OrderHandler) Get(c *gin.Context) {
	orderNo := c.Param("order_no")
	if orderNo == "" {
		response.ErrorWithCode(c, apperror.CodeMissingParam)
		return
	}

	order, err := h.orderUC.Get(c.Request.Context(), orderNo)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, toOrderResponse(order))
}

func (h *OrderHandler) Close(c *gin.Context) {
	orderNo := c.Param("order_no")
	if orderNo == "" {
		response.ErrorWithCode(c, apperror.CodeMissingParam)
		return
	}

	if err := h.orderUC.Close(c.Request.Context(), orderNo); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, dto.Empty{})
}

func toOrderResponse(order *entity.Order) dto.OrderResponse {
	resp := dto.OrderResponse{
		OrderNo:     order.OrderNo,
		UserID:      order.UserID,
		Subject:     order.Subject,
		Amount:      order.Amount,
		Currency:    order.Currency,
		Status:      order.Status.String(),
		ExpiredAt:   order.ExpiredAt.Unix(),
		Description: order.Description,
		Metadata:    order.Metadata,
		CreatedAt:   order.CreatedAt.Unix(),
	}
	if order.PaidAt != nil {
		t := order.PaidAt.Unix()
		resp.PaidAt = &t
	}
	return resp
}
