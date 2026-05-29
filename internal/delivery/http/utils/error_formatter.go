package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func FormatValidationError(err error) map[string]string {
	errs := make(map[string]string)

	// 1. Handle Type Mismatch Errors (e.g., sending a string for an int)
	if unmarshalErr, ok := err.(*json.UnmarshalTypeError); ok {
		// errs[unmarshalErr.Field] = fmt.Sprintf("invalid type: expected %s but got %s", unmarshalErr.Type.String(), unmarshalErr.Value)
		errs[unmarshalErr.Field] = fmt.Sprintf("invalid type for field %s", unmarshalErr.Field)
		return errs
	}

	// 2. Handle Validation Errors (e.g., required fields, min/max lengths)
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			field := e.Field()

			// Determine pluralization based on the Kind of the field
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
				if strings.Contains(strings.ToLower(e.Param()), "password") || strings.Contains(strings.ToLower(field), "password") {
					errs[field] = "passwords do not match"
				} else {
					errs[field] = fmt.Sprintf("%s must be equal to %s", field, e.Param())
				}
			default:
				errs[field] = fmt.Sprintf("%s is invalid", field)
			}
		}
	}

	return errs
}
