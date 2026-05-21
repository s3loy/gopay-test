package validator

import (
	"testing"

	"github.com/s3loy/gopay/internal/pkg/apperror"
)

type testStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Age   int    `validate:"gte=0,lte=150"`
}

func TestValidateStruct_Success(t *testing.T) {
	s := testStruct{Name: "test", Email: "test@example.com", Age: 25}
	if err := ValidateStruct(&s); err != nil {
		t.Errorf("ValidateStruct() error = %v", err)
	}
}

func TestValidateStruct_MissingRequired(t *testing.T) {
	s := testStruct{Email: "test@example.com", Age: 25}
	err := ValidateStruct(&s)
	if err == nil {
		t.Fatal("expected error for missing required field")
	}
	if !apperror.Is(err, apperror.CodeInvalidParams) {
		t.Errorf("error code mismatch: %v", err)
	}
}

func TestValidateStruct_InvalidEmail(t *testing.T) {
	s := testStruct{Name: "test", Email: "not-an-email", Age: 25}
	err := ValidateStruct(&s)
	if err == nil {
		t.Fatal("expected error for invalid email")
	}
	if !apperror.Is(err, apperror.CodeInvalidParams) {
		t.Errorf("error code mismatch: %v", err)
	}
}

func TestValidateStruct_AgeOutOfRange(t *testing.T) {
	s := testStruct{Name: "test", Email: "test@example.com", Age: 200}
	err := ValidateStruct(&s)
	if err == nil {
		t.Fatal("expected error for age out of range")
	}
	if !apperror.Is(err, apperror.CodeInvalidParams) {
		t.Errorf("error code mismatch: %v", err)
	}
}

func TestValidateStruct_InvalidInputType(t *testing.T) {
	err := ValidateStruct("not a struct")
	if err == nil {
		t.Fatal("expected error for non-struct input")
	}
	if !apperror.Is(err, apperror.CodeInvalidParams) {
		t.Errorf("error code mismatch: %v", err)
	}
}

func TestV(t *testing.T) {
	v := V()
	if v == nil {
		t.Error("V() should not return nil")
	}
}
