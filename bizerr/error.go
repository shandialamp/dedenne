package bizerr

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v5"
	"github.com/shandialamp/dedenne/response"
	"go.uber.org/zap"
)

// ErrorCode 业务错误码
type ErrorCode int

const (
	// 通用错误码
	CodeValidationError         ErrorCode = 10001 // 参数验证错误
	CodeNotFound                ErrorCode = 10002 // 资源不存在
	CodeAlreadyExists           ErrorCode = 10003 // 资源已存在
	CodeUnauthorized            ErrorCode = 10004 // 未授权
	CodeForbidden               ErrorCode = 10005 // 禁止访问
	CodeInternalError           ErrorCode = 10006 // 内部服务错误
	CodeInvalidToken            ErrorCode = 10007 // token 无效
	CodeTokenExpired            ErrorCode = 10008 // token 已过期
	CodeInsufficientPermissions ErrorCode = 10009 // 权限不足
)

// BusinessError 业务错误（不需要记录错误日志）
type BusinessError struct {
	Code    ErrorCode // 错误码
	Message string    // 用户可见的错误消息
	Details string    // 错误详情（可选）
}

func (e *BusinessError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("code=%d, message=%s, details=%s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("code=%d, message=%s", e.Code, e.Message)
}

// NewBusinessError 创建业务错误
func NewBusinessError(code ErrorCode, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Details: "",
	}
}

// NewBusinessErrorWithDetails 创建带详情的业务错误
func NewBusinessErrorWithDetails(code ErrorCode, message, details string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// IsBusinessError 判断是否是业务错误
func IsBusinessError(err error) bool {
	_, ok := err.(*BusinessError)
	return ok
}

// ValidationError 参数验证错误
func ValidationError(message string) *BusinessError {
	return NewBusinessError(CodeValidationError, message)
}

// NotFoundError 资源不存在错误
func NotFoundError(resource string) *BusinessError {
	return NewBusinessError(CodeNotFound, fmt.Sprintf("%s not found", resource))
}

// AlreadyExistsError 资源已存在错误
func AlreadyExistsError(resource string) *BusinessError {
	return NewBusinessError(CodeAlreadyExists, fmt.Sprintf("%s already exists", resource))
}

// UnauthorizedError 未授权错误
func UnauthorizedError() *BusinessError {
	return NewBusinessError(CodeUnauthorized, "Unauthorized")
}

// ForbiddenError 禁止访问错误
func ForbiddenError() *BusinessError {
	return NewBusinessError(CodeForbidden, "Forbidden")
}

// InvalidTokenError token 无效错误
func InvalidTokenError() *BusinessError {
	return NewBusinessError(CodeInvalidToken, "Invalid token")
}

// TokenExpiredError token 已过期错误
func TokenExpiredError() *BusinessError {
	return NewBusinessError(CodeTokenExpired, "Token expired")
}

// InternalError 内部服务错误
func InternalError(message string) error {
	return fmt.Errorf("internal error: %s", message)
}

// HTTPErrorHandler 统一处理 Echo 的错误响应
func HTTPErrorHandler(logger *zap.Logger) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {
		code := http.StatusInternalServerError
		message := "Internal Server Error"
		httpCode := http.StatusInternalServerError

		// 优先处理业务错误 - 记录 INFO 级别
		if be, ok := err.(*BusinessError); ok {
			code = int(be.Code)
			message = be.Message
			httpCode = http.StatusOK
			resp := &response.Response{
				Code:    code,
				Data:    nil,
				Message: message,
				Error:   be.Details,
			}
			logger.Info("Business error",
				zap.Int("code", code),
				zap.String("message", message),
				zap.String("path", c.Request().URL.Path),
				zap.String("method", c.Request().Method),
			)
			c.JSON(httpCode, resp)
			return
		}

		// 处理 Echo HTTP 错误 - 记录 WARN
		if he, ok := err.(*echo.HTTPError); ok {
			httpCode = he.Code
			message = he.Message
			logger.Warn("HTTP error",
				zap.Int("code", httpCode),
				zap.String("message", message),
				zap.String("path", c.Request().URL.Path),
				zap.String("method", c.Request().Method),
			)
			resp := &response.Response{
				Code:    httpCode,
				Data:    nil,
				Message: message,
				Error:   "",
			}
			c.JSON(httpCode, resp)
			return
		}

		// 处理 404 Not Found（Echo 直接返回的字符串错误）
		if err.Error() == "Not Found" {
			httpCode = http.StatusNotFound
			message = "Not Found"
			logger.Warn("HTTP error",
				zap.Int("code", httpCode),
				zap.String("message", message),
				zap.String("path", c.Request().URL.Path),
				zap.String("method", c.Request().Method),
			)
			resp := &response.Response{
				Code:    httpCode,
				Data:    nil,
				Message: message,
				Error:   "",
			}
			c.JSON(httpCode, resp)
			return
		}

		// 处理其他系统错误 - 记录 ERROR + goroutine 堆栈
		// 注意: zap.Stack("stacktrace") 只记录 HTTPErrorHandler 自身的调用栈，
		// 无法追溯到原始 handler 中发生错误的位置。
		// 因此额外记录 goroutine 堆栈，帮助定位原始错误来源。
		logger.Error("Internal server error",
			zap.String("path", c.Request().URL.Path),
			zap.String("method", c.Request().Method),
			zap.Error(err),
			zap.String("goroutine_stack", string(debug.Stack())),
		)
		resp := &response.Response{
			Code:    500,
			Data:    nil,
			Message: message,
			Error:   "",
		}
		c.JSON(httpCode, resp)
	}
}
