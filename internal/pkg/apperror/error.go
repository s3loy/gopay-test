package apperror

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code       int                    `json:"code"`
	Message    string                 `json:"message"`
	HTTPStatus int                    `json:"-"`
	Cause      error                  `json:"-"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func New(code int, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
	}
}

func NewWithStatus(code int, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
	}
}

func Wrap(err error, code int) *AppError {
	if err == nil {
		return nil
	}
	if ae, ok := err.(*AppError); ok {
		return ae
	}
	return &AppError{
		Code:       code,
		Message:    GetMessage(code),
		HTTPStatus: GetHTTPStatus(code),
		Cause:      err,
	}
}

func Is(err error, code int) bool {
	if err == nil {
		return false
	}
	if ae, ok := err.(*AppError); ok {
		return ae.Code == code
	}
	return false
}

func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

func (e *AppError) WithCause(err error) *AppError {
	e.Cause = err
	return e
}

func (e *AppError) WithHTTPStatus(status int) *AppError {
	e.HTTPStatus = status
	return e
}
