package apperror

import (
	"errors"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(CodeOrderNotFound, "order not found")
	if err.Code != CodeOrderNotFound {
		t.Errorf("Code = %d, want %d", err.Code, CodeOrderNotFound)
	}
	if err.Message != "order not found" {
		t.Errorf("Message = %s, want %s", err.Message, "order not found")
	}
	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("HTTPStatus = %d, want %d", err.HTTPStatus, http.StatusNotFound)
	}
}

func TestWrap(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		if got := Wrap(nil, CodeDatabaseError); got != nil {
			t.Error("Wrap(nil) should return nil")
		}
	})

	t.Run("wrap standard error", func(t *testing.T) {
		inner := errors.New("connection refused")
		err := Wrap(inner, CodeDatabaseError)
		if err == nil {
			t.Fatal("Wrap should return non-nil")
		}
		if err.Code != CodeDatabaseError {
			t.Errorf("Code = %d, want %d", err.Code, CodeDatabaseError)
		}
		if err.Cause != inner {
			t.Error("Cause should be the inner error")
		}
	})

	t.Run("wrap AppError returns original", func(t *testing.T) {
		original := New(CodeOrderNotFound, "not found")
		err := Wrap(original, CodeDatabaseError)
		if err != original {
			t.Error("Wrap(AppError) should return the original AppError")
		}
	})
}

func TestIs(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		err := New(CodeOrderNotFound, "not found")
		if !Is(err, CodeOrderNotFound) {
			t.Error("Is should return true for matching code")
		}
	})

	t.Run("no match", func(t *testing.T) {
		err := New(CodeOrderNotFound, "not found")
		if Is(err, CodePaymentNotFound) {
			t.Error("Is should return false for non-matching code")
		}
	})

	t.Run("nil error", func(t *testing.T) {
		if Is(nil, CodeOrderNotFound) {
			t.Error("Is(nil) should return false")
		}
	})

	t.Run("non-AppError", func(t *testing.T) {
		err := errors.New("some error")
		if Is(err, CodeOrderNotFound) {
			t.Error("Is should return false for non-AppError")
		}
	})
}

func TestAppError_Error(t *testing.T) {
	t.Run("with cause", func(t *testing.T) {
		err := Wrap(errors.New("inner"), CodeDatabaseError)
		want := "[100400] database error: inner"
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("without cause", func(t *testing.T) {
		err := New(CodeOrderNotFound, "not found")
		want := "[101001] not found"
		if got := err.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})
}

func TestAppError_WithDetail(t *testing.T) {
	err := New(CodeOrderNotFound, "not found").
		WithDetail("order_no", "ORD123").
		WithDetail("user_id", 42)

	if err.Details["order_no"] != "ORD123" {
		t.Error("Detail order_no mismatch")
	}
	if err.Details["user_id"] != 42 {
		t.Error("Detail user_id mismatch")
	}
}

func TestAppError_WithCause(t *testing.T) {
	inner := errors.New("inner")
	err := New(CodeDatabaseError, "db error").WithCause(inner)
	if err.Cause != inner {
		t.Error("Cause mismatch")
	}
}

func TestAppError_WithHTTPStatus(t *testing.T) {
	err := New(CodeOrderNotFound, "not found").WithHTTPStatus(http.StatusNotFound)
	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("HTTPStatus = %d, want %d", err.HTTPStatus, http.StatusNotFound)
	}
}

func TestUnwrap(t *testing.T) {
	inner := errors.New("inner")
	err := Wrap(inner, CodeDatabaseError)
	if err.Unwrap() != inner {
		t.Error("Unwrap should return the cause")
	}
}
