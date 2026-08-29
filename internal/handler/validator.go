package handler

import "github.com/go-playground/validator/v10"

// Validator adapts go-playground/validator to echo.Validator so that the
// `validate:` tags on request structs are enforced by c.Validate.
type Validator struct {
	v *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{v: validator.New()}
}

func (cv *Validator) Validate(i interface{}) error {
	return cv.v.Struct(i)
}
