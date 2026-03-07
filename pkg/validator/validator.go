package validator

import (
	"github.com/go-playground/validator/v10"
)

// Validator wraps go-playground/validator for shared usage across services.
type Validator struct {
	Validate *validator.Validate
}

// New creates a new Validator instance.
func New() *Validator {
	return &Validator{
		Validate: validator.New(),
	}
}

// Struct validates a struct and returns an error if validation fails.
func (v *Validator) Struct(s interface{}) error {
	return v.Validate.Struct(s)
}
