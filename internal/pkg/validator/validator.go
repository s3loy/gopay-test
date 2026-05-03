package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/s3loy/gopay/internal/pkg/apperror"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func V() *validator.Validate {
	return validate
}

func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		if verrs, ok := err.(validator.ValidationErrors); ok {
			return apperror.New(apperror.CodeInvalidParams, verrs[0].Error())
		}
		return apperror.New(apperror.CodeInvalidParams, err.Error())
	}
	return nil
}
