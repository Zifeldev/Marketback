package apperrors

import (
	"errors"
	"fmt"
	"net/http"
)

// Error codes
const (
	CodeNotFound          = "NOT_FOUND"
	CodeBadRequest        = "BAD_REQUEST"
	CodeUnauthorized      = "UNAUTHORIZED"
	CodeForbidden         = "FORBIDDEN"
	CodeConflict          = "CONFLICT"
	CodeInternalError     = "INTERNAL_ERROR"
	CodeValidationError   = "VALIDATION_ERROR"
	CodeInsufficientStock = "INSUFFICIENT_STOCK"
	CodeEmptyCart         = "EMPTY_CART"
)

// AppError represents a typed application error
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Wrap wraps an existing error with AppError
func Wrap(err error, code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// Common errors
var (
	ErrNotFound          = New(CodeNotFound, "resource not found", http.StatusNotFound)
	ErrBadRequest        = New(CodeBadRequest, "bad request", http.StatusBadRequest)
	ErrUnauthorized      = New(CodeUnauthorized, "unauthorized", http.StatusUnauthorized)
	ErrForbidden         = New(CodeForbidden, "forbidden", http.StatusForbidden)
	ErrInternalError     = New(CodeInternalError, "internal server error", http.StatusInternalServerError)
	ErrInsufficientStock = New(CodeInsufficientStock, "insufficient stock", http.StatusConflict)
	ErrEmptyCart         = New(CodeEmptyCart, "cart is empty", http.StatusBadRequest)
)

// Entity-specific not found errors
func ProductNotFound(id int) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    fmt.Sprintf("product with id %d not found", id),
		HTTPStatus: http.StatusNotFound,
	}
}

func CategoryNotFound(id int) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    fmt.Sprintf("category with id %d not found", id),
		HTTPStatus: http.StatusNotFound,
	}
}

func OrderNotFound(id int) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    fmt.Sprintf("order with id %d not found", id),
		HTTPStatus: http.StatusNotFound,
	}
}

func SellerNotFound(id int) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    fmt.Sprintf("seller with id %d not found", id),
		HTTPStatus: http.StatusNotFound,
	}
}

func CartItemNotFound(id int) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    fmt.Sprintf("cart item with id %d not found", id),
		HTTPStatus: http.StatusNotFound,
	}
}

// ValidationError creates validation error
func ValidationError(field, message string) *AppError {
	return &AppError{
		Code:       CodeValidationError,
		Message:    fmt.Sprintf("validation error: %s - %s", field, message),
		HTTPStatus: http.StatusBadRequest,
	}
}

// Conflict creates conflict error
func Conflict(message string) *AppError {
	return &AppError{
		Code:       CodeConflict,
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

// InsufficientStockForProduct creates an error for a specific product
func InsufficientStockForProduct(productID int) *AppError {
	return &AppError{
		Code:       CodeInsufficientStock,
		Message:    fmt.Sprintf("insufficient stock for product %d", productID),
		HTTPStatus: http.StatusConflict,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from error chain
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// GetHTTPStatus returns HTTP status code for an error
func GetHTTPStatus(err error) int {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}
