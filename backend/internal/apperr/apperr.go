package apperr

import "net/http"

// AppError はアプリケーション層で扱う HTTP ステータス付きのエラーです。
type AppError struct {
	Status  int    `json:"status"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// Error はエラーメッセージを返します。
func (e *AppError) Error() string {
	return e.Message
}

// New は新しい AppError を作成します。
func New(status int, message string) *AppError {
	return &AppError{
		Status:  status,
		Code:    defaultCode(status),
		Message: message,
	}
}

// NewWithCode は code 付きの AppError を作成します。
func NewWithCode(status int, code, message string) *AppError {
	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

// API 共通エラー
var (
	ErrBadRequest = NewWithCode(http.StatusBadRequest, "bad_request", "bad request")
	ErrForbidden  = NewWithCode(http.StatusForbidden, "forbidden", "forbidden")
	ErrNotFound   = NewWithCode(http.StatusNotFound, "not_found", "not found")
)

// NotFound はリソース名付きの 404 エラーを返します。
func NotFound(resource string) *AppError {
	return NewWithCode(http.StatusNotFound, "not_found", resource+" not found")
}

// BadRequest はメッセージ付きの 400 エラーを返します。
func BadRequest(message string) *AppError {
	return NewWithCode(http.StatusBadRequest, "bad_request", message)
}

// BadRequestWithCode は code とメッセージ付きの 400 エラーを返します。
func BadRequestWithCode(code, message string) *AppError {
	return NewWithCode(http.StatusBadRequest, code, message)
}

// InternalError はメッセージ付きの 500 エラーを返します。
func InternalError(message string) *AppError {
	return NewWithCode(http.StatusInternalServerError, "internal_error", message)
}

func defaultCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	default:
		if status >= http.StatusInternalServerError {
			return "internal_error"
		}
		return ""
	}
}
