package web

import "github.com/go-playground/validator/v10"

func NewValidator() *validator.Validate {
	validator := validator.New(validator.WithRequiredStructEnabled())

	return validator
}
