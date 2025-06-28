package errors

import "fmt"

// NotFoundError represents an error when a resource is not found.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("%s with ID '%s' not found", e.Resource, e.ID)
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

// NewNotFoundError creates a new NotFoundError.
func NewNotFoundError(resource string, id string) *NotFoundError {
	return &NotFoundError{Resource: resource, ID: id}
}

// ValidationError represents an error during data validation.
type ValidationError struct {
	Message string
	Field   string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation failed: %s", e.Message)
}

// NewValidationError creates a new ValidationError.
func NewValidationError(message string, field string) *ValidationError {
	return &ValidationError{Message: message, Field: field}
}

// UnauthorizedError represents an authentication or authorization failure.
type UnauthorizedError struct {
	Message string
}

func (e *UnauthorizedError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("unauthorized: %s", e.Message)
	}
	return "unauthorized"
}

// NewUnauthorizedError creates a new UnauthorizedError.
func NewUnauthorizedError(message string) *UnauthorizedError {
	return &UnauthorizedError{Message: message}
}

// ForbiddenError represents a permission denied error.
type ForbiddenError struct {
    Message string
}

func (e *ForbiddenError) Error() string {
    if e.Message != "" {
        return fmt.Sprintf("forbidden: %s", e.Message)
    }
    return "forbidden"
}

// NewForbiddenError creates a new ForbiddenError.
func NewForbiddenError(message string) *ForbiddenError {
    return &ForbiddenError{Message: message}
}


// ConflictError represents an error due to a conflict with the current state of the resource.
type ConflictError struct {
    Message string
}

func (e *ConflictError) Error() string {
    if e.Message != "" {
        return fmt.Sprintf("conflict: %s", e.Message)
    }
    return "conflict"
}

// NewConflictError creates a new ConflictError.
func NewConflictError(message string) *ConflictError {
    return &ConflictError{Message: message}
}

// InternalServerError represents a generic server error.
type InternalServerError struct {
    Message string
    Err     error // Underlying error
}

func (e *InternalServerError) Error() string {
    if e.Message == "" && e.Err != nil {
        return fmt.Sprintf("internal server error: %v", e.Err)
    }
    if e.Message != "" && e.Err != nil {
        return fmt.Sprintf("internal server error: %s - %v", e.Message, e.Err)
    }
    if e.Message != "" {
        return fmt.Sprintf("internal server error: %s", e.Message)
    }
    return "internal server error"
}

func (e *InternalServerError) Unwrap() error {
    return e.Err
}

// NewInternalServerError creates a new InternalServerError.
func NewInternalServerError(message string, err error) *InternalServerError {
    return &InternalServerError{Message: message, Err: err}
}
