package main

import "fmt"

type ErrorCode int

const (
	CodeSuccess ErrorCode = 200

	CodeInvalidParams ErrorCode = 400
	CodeNotReady      ErrorCode = 503
	CodeTransFailure  ErrorCode = 500
	CodeInternalError ErrorCode = 502
)

type StandardError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e *StandardError) Error() string {
	return fmt.Sprintf("code=%d, message=%s", e.Code, e.Message)
}

func NewError(code ErrorCode, message string) *StandardError {
	return &StandardError{
		Code:    code,
		Message: message,
	}
}

type StandardResponse struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func NewSuccessResponse(data interface{}) *StandardResponse {
	return &StandardResponse{
		Code: CodeSuccess,
		Data: data,
	}
}

func NewErrorResponse(code ErrorCode, message string) *StandardResponse {
	return &StandardResponse{
		Code:    code,
		Message: message,
	}
}
