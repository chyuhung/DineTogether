package errors

import "net/http"

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// 常见错误类型
var (
	ErrBadRequest     = NewAppError(http.StatusBadRequest, "无效的请求")
	ErrUnauthorized   = NewAppError(http.StatusUnauthorized, "未授权")
	ErrNotFound       = NewAppError(http.StatusNotFound, "资源未找到")
	ErrInternalServer = NewAppError(http.StatusInternalServerError, "服务器错误")
)
