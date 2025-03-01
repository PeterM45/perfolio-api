package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator defines validation methods
type Validator interface {
	Validate(i interface{}) error
}

// CustomValidator implements Validator with validator.v10
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator
func NewValidator() Validator {
	v := validator.New()

	// Register a function to get the field name from the struct tags
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &CustomValidator{
		validator: v,
	}
}

// Validate validates a struct
func (cv *CustomValidator) Validate(i interface{}) error {
	err := cv.validator.Struct(i)
	if err == nil {
		return nil
	}

	// Convert validation errors to a friendly format
	validationErrors := err.(validator.ValidationErrors)
	errorMessages := make([]string, 0, len(validationErrors))

	for _, e := range validationErrors {
		errorMessages = append(errorMessages, formatError(e))
	}

	return fmt.Errorf("validation failed: %s", strings.Join(errorMessages, "; "))
}

// formatError formats a validation error
func formatError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, param)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	default:
		return fmt.Sprintf("%s failed validation: %s=%s", field, tag, param)
	}
}
