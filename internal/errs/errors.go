package errs

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("conflict")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrValidation      = ErrInvalidArgument
	ErrInternal        = errors.New("internal")
)

type AppError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
	Err     error             `json:"-"`
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Code
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewNotFound(message string) *AppError {
	return &AppError{Code: "not_found", Message: message, Err: ErrNotFound}
}

func NewConflict(message string) *AppError {
	return &AppError{Code: "conflict", Message: message, Err: ErrConflict}
}

func NewInvalidArgument(message string, fields map[string]string) *AppError {
	return &AppError{
		Code:    "validation_error",
		Message: message,
		Fields:  fields,
		Err:     ErrInvalidArgument,
	}
}

func NewValidation(message string, fields map[string]string) *AppError {
	return NewInvalidArgument(message, fields)
}

func NewInternal(err error) *AppError {
	return &AppError{
		Code:    "internal_error",
		Message: "internal server error",
		Err:     errors.Join(ErrInternal, err),
	}
}
