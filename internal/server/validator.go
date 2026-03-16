package server

import "github.com/go-playground/validator/v10"

type CustomValidator struct {
	validate *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	return &CustomValidator{validate: validator.New()}
}

func (v *CustomValidator) Validate(i any) error {
	return v.validate.Struct(i)
}
