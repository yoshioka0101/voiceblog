package apperr

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
