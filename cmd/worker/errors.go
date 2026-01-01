package main

import "fmt"

// ErrorCode represents standardized error codes
type ErrorCode int

const (
	// Success code
	CodeSuccess ErrorCode = 200
	// Load error codes (1000-1099)
	CodeLoadInvalidParams   ErrorCode = 1000
	CodeLoadPathNotExists   ErrorCode = 1001
	CodeLoadIncompleteFiles ErrorCode = 1002
	CodeLoadInternalError   ErrorCode = 1003
	CodeLoadUnknownError    ErrorCode = 1009
	// Exit error codes (1100-1199)
	CodeExitInvalidParams ErrorCode = 1100
	CodeExitWaitingTask   ErrorCode = 1101
	CodeExitInternalError ErrorCode = 1109
	// Trans error codes (1200-1299)
	CodeTransInvalidParams ErrorCode = 1200
	CodeTransFailure       ErrorCode = 1201
	CodeTransInternalError ErrorCode = 1209
)

// StandardError represents a standard error response
type StandardError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

// Error implements the error interface
func (e *StandardError) Error() string {
	return fmt.Sprintf("code=%d, message=%s", e.Code, e.Message)
}

// NewError creates a new StandardError
func NewError(code ErrorCode, message string) *StandardError {
	return &StandardError{
		Code:    code,
		Message: message,
	}
}

// StandardResponse represents a standard response with data
type StandardResponse struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// NewSuccessResponse creates a success response
func NewSuccessResponse(data interface{}) *StandardResponse {
	return &StandardResponse{
		Code: CodeSuccess,
		Data: data,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(code ErrorCode, message string) *StandardResponse {
	return &StandardResponse{
		Code:    code,
		Message: message,
	}
}
