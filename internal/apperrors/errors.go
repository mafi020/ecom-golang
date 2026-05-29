package apperrors

import "fmt"

type NotFoundError struct {
	Resource string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

type ConflictError struct {
	Errors map[string]string `json:"errors"`
}

func (e *ConflictError) Error() string {
	return "conflict detected"
}

type ValidationError struct {
	Errors map[string]string `json:"errors"`
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

type BadRequestError struct {
	Errors map[string]string `json:"errors"`
}

func (e *BadRequestError) Error() string {
	return "bad request"
}

type UnauthorizedError struct {
	Message string `json:"message"`
}

func (e *UnauthorizedError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "unauthorized"
}

type ForbiddenError struct {
	Message string `json:"message"`
}

func (e *ForbiddenError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "forbidden"
}

type TooManyRequestsError struct {
	Message string `json:"message"`
}

func (e *TooManyRequestsError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "too many requests"
}
