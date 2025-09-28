package main

import "fmt"

// ErrorCode represents standardized error codes
type ErrorCode int

const (
	// Success code
	CodeSuccess ErrorCode = 0

	// Poweron error codes (1000-1099)
	CodePoweronInvalidParams   ErrorCode = 1000
	CodePoweronPathNotExists   ErrorCode = 1001
	CodePoweronIncompleteFiles ErrorCode = 1002
	CodePoweronInternalError   ErrorCode = 1003
	CodePoweronUnknownError    ErrorCode = 1009

	// Poweroff error codes (1100-1199)
	CodePoweroffInvalidParams ErrorCode = 1100
	CodePoweroffWaitingTask   ErrorCode = 1101
	CodePoweroffInternalError ErrorCode = 1109

	// Reboot error codes (1150-1159)
	CodeRebootInvalidParams ErrorCode = 1150
	CodeRebootInternalError ErrorCode = 1151
	CodeRebootWaitingTask   ErrorCode = 1152

	// Compute error codes (1200-1299)
	CodeComputeInvalidParams ErrorCode = 1200
	CodeComputeFailure       ErrorCode = 1201
	CodeComputeInternalError ErrorCode = 1209
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
