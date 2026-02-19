package forge

import "fmt"

// Error represents a structured application error with HTTP status code mapping.
// Both API (JSON) and HTML (SSE) handlers use this type for consistent error responses.
type Error struct {
	// Status is the HTTP status code (e.g. 400, 404, 500).
	Status int

	// Code is a machine-readable error code (e.g. "not_found", "validation_error").
	Code string

	// Message is a human-readable error summary.
	Message string

	// Detail provides additional context for debugging (may be empty).
	Detail string

	// Err is the underlying error, if any.
	Err error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error for errors.Is/errors.As compatibility.
func (e *Error) Unwrap() error {
	return e.Err
}

// GetStatus returns the HTTP status code.
func (e *Error) GetStatus() int {
	return e.Status
}

// NotFound returns a 404 error for a missing resource.
func NotFound(resource string, id any) *Error {
	return &Error{
		Status:  404,
		Code:    "resource_not_found",
		Message: fmt.Sprintf("%s not found", resource),
		Detail:  fmt.Sprintf("ID: %v", id),
	}
}

// UniqueViolation returns a 409 error for unique constraint violations.
func UniqueViolation(field string) *Error {
	return &Error{
		Status:  409,
		Code:    "unique_violation",
		Message: fmt.Sprintf("%s already exists", field),
		Detail:  field,
	}
}

// ForeignKeyViolation returns a 400 error for foreign key constraint violations.
func ForeignKeyViolation(field string) *Error {
	return &Error{
		Status:  400,
		Code:    "foreign_key_violation",
		Message: "Referenced resource does not exist",
		Detail:  field,
	}
}

// Unauthorized returns a 401 error for authentication failures.
func Unauthorized(message string) *Error {
	return &Error{
		Status:  401,
		Code:    "unauthorized",
		Message: message,
	}
}

// Forbidden returns a 403 error for authorization failures.
func Forbidden(message string) *Error {
	return &Error{
		Status:  403,
		Code:    "forbidden",
		Message: message,
	}
}

// BadRequest returns a 400 error for invalid requests.
func BadRequest(message string) *Error {
	return &Error{
		Status:  400,
		Code:    "bad_request",
		Message: message,
	}
}

// InternalError returns a 500 error wrapping an unexpected error.
func InternalError(err error) *Error {
	return &Error{
		Status:  500,
		Code:    "internal_error",
		Message: "An unexpected error occurred",
		Err:     err,
	}
}

// NewValidationError creates a 422 error from validation errors.
// Accepts any type with an Error() method to avoid importing validation package.
func NewValidationError(errs interface{}) *Error {
	var detail string
	if e, ok := errs.(error); ok {
		detail = e.Error()
	} else {
		detail = fmt.Sprintf("%v", errs)
	}

	return &Error{
		Status:  422,
		Code:    "validation_error",
		Message: "Validation failed",
		Detail:  detail,
	}
}
