package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/s3loy/gopay/internal/pkg/apperror"
)

type Response struct {
	Code      int                    `json:"code"`
	Message   string                 `json:"message"`
	Data      interface{}            `json:"data,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func JSON(c *gin.Context, code int, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      code,
		Message:   apperror.GetMessage(code),
		Data:      data,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().Unix(),
	})
}

func OK(c *gin.Context, data interface{}) {
	JSON(c, apperror.CodeSuccess, data)
}

func Page(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	OK(c, PageData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func Error(c *gin.Context, err error) {
	if ae, ok := err.(*apperror.AppError); ok {
		c.JSON(ae.HTTPStatus, Response{
			Code:      ae.Code,
			Message:   ae.Message,
			Details:   ae.Details,
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now().Unix(),
		})
		return
	}

	c.JSON(http.StatusInternalServerError, Response{
		Code:      apperror.CodeUnknown,
		Message:   err.Error(),
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().Unix(),
	})
}

func ErrorWithCode(c *gin.Context, code int) {
	c.JSON(apperror.GetHTTPStatus(code), Response{
		Code:      code,
		Message:   apperror.GetMessage(code),
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().Unix(),
	})
}
