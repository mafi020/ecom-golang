package apperrors

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func ParseValidationError(err error) *ValidationError {
	errs := make(map[string]string)

	if unmarshalErr, ok := err.(*json.UnmarshalTypeError); ok {
		errs[unmarshalErr.Field] = fmt.Sprintf("invalid type for field %s", unmarshalErr.Field)
		return &ValidationError{Errors: errs}
	}

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			field := e.Field()

			verb := "is"
			if e.Kind() == reflect.Slice || e.Kind() == reflect.Array || e.Kind() == reflect.Map {
				verb = "are"
			}

			switch e.Tag() {
			case "required":
				errs[field] = fmt.Sprintf("%s %s required", field, verb)
			case "email":
				errs[field] = "invalid email format"
			case "lt":
				errs[field] = fmt.Sprintf("%s must be less than %s", field, e.Param())
			case "lte":
				errs[field] = fmt.Sprintf("%s must be less than or equal to %s", field, e.Param())
			case "gt":
				errs[field] = fmt.Sprintf("%s must be greater than %s", field, e.Param())
			case "gte":
				errs[field] = fmt.Sprintf("%s must be greater than or equal to %s", field, e.Param())
			case "min":
				errs[field] = fmt.Sprintf("%s must be at least %s characters", field, e.Param())
			case "max":
				errs[field] = fmt.Sprintf("%s must be at most %s characters", field, e.Param())
			case "eqfield":
				if strings.Contains(strings.ToLower(e.Param()), "password") ||
					strings.Contains(strings.ToLower(field), "password") {
					errs[field] = "passwords do not match"
				} else {
					errs[field] = fmt.Sprintf("%s must be equal to %s", field, e.Param())
				}
			default:
				errs[field] = fmt.Sprintf("%s is invalid", field)
			}
		}
	}

	return &ValidationError{Errors: errs}
}
