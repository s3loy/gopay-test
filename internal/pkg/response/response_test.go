package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/s3loy/gopay/internal/pkg/apperror"
)

func setupGin() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	return c, w
}

func TestOK(t *testing.T) {
	c, w := setupGin()
	OK(c, gin.H{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != apperror.CodeSuccess {
		t.Errorf("code = %d, want %d", resp.Code, apperror.CodeSuccess)
	}
	if resp.Message != "success" {
		t.Errorf("message = %v, want success", resp.Message)
	}
}

func TestPage(t *testing.T) {
	c, w := setupGin()
	Page(c, []string{"a", "b"}, 10, 1, 20)

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data type = %T, want map", resp.Data)
	}
	if data["total"] != float64(10) {
		t.Errorf("total = %v, want 10", data["total"])
	}
	if data["page"] != float64(1) {
		t.Errorf("page = %v, want 1", data["page"])
	}
	if data["page_size"] != float64(20) {
		t.Errorf("page_size = %v, want 20", data["page_size"])
	}
}

func TestError_WithAppError(t *testing.T) {
	c, w := setupGin()
	appErr := apperror.New(apperror.CodeOrderNotFound, "order not found").WithHTTPStatus(http.StatusNotFound)
	Error(c, appErr)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != apperror.CodeOrderNotFound {
		t.Errorf("code = %d, want %d", resp.Code, apperror.CodeOrderNotFound)
	}
	if resp.Message != "order not found" {
		t.Errorf("message = %v, want order not found", resp.Message)
	}
}

func TestError_WithPlainError(t *testing.T) {
	c, w := setupGin()
	Error(c, errors.New("plain error"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != apperror.CodeUnknown {
		t.Errorf("code = %d, want %d", resp.Code, apperror.CodeUnknown)
	}
}

func TestErrorWithCode(t *testing.T) {
	c, w := setupGin()
	ErrorWithCode(c, apperror.CodeInvalidParams)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != apperror.CodeInvalidParams {
		t.Errorf("code = %d, want %d", resp.Code, apperror.CodeInvalidParams)
	}
	if resp.Message != apperror.GetMessage(apperror.CodeInvalidParams) {
		t.Errorf("message = %v, want %v", resp.Message, apperror.GetMessage(apperror.CodeInvalidParams))
	}
}

func TestJSON(t *testing.T) {
	c, w := setupGin()
	c.Set("request_id", "req-123")
	JSON(c, apperror.CodeSuccess, gin.H{"data": "test"})

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.RequestID != "req-123" {
		t.Errorf("request_id = %v, want req-123", resp.RequestID)
	}
	if resp.Timestamp == 0 {
		t.Error("timestamp should not be zero")
	}
}
