package apperr

import "net/http"

// AppError はアプリケーション層で扱う HTTP ステータス付きのエラーです。
type AppError struct {
	Status  int    `json:"status"`
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
		Message: message,
	}
}

// API 共通エラー
var (
	ErrBadRequest = New(http.StatusBadRequest, "bad request")
	ErrForbidden  = New(http.StatusForbidden, "forbidden")
	ErrNotFound   = New(http.StatusNotFound, "not found")
)

// NotFound はリソース名付きの 404 エラーを返します。
func NotFound(resource string) *AppError {
	return New(http.StatusNotFound, resource+" not found")
}

// BadRequest はメッセージ付きの 400 エラーを返します。
func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, message)
}

// InternalError はメッセージ付きの 500 エラーを返します。
func InternalError(message string) *AppError {
	return New(http.StatusInternalServerError, message)
}
