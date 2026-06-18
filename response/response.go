package response

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"`    // 0 表示成功，非 0 表示业务错误或系统错误
	Data    interface{} `json:"data"`    // 响应数据
	Message string      `json:"message"` // 消息/错误描述
	Error   string      `json:"error"`   // 错误详情（可选）
}

// Success 成功响应
// nolint:vet // Echo v5 Context contains sync.RWMutex (expected)
func Success(c *echo.Context, data interface{}, message string) error {
	resp := &Response{
		Code:    0,
		Data:    data,
		Message: message,
		Error:   "",
	}
	return c.JSON(http.StatusOK, resp)
}

// SuccessWithCode 自定义成功响应（带状态码）
func SuccessWithCode(c *echo.Context, statusCode int, data interface{}, message string) error {
	resp := &Response{
		Code:    0,
		Data:    data,
		Message: message,
		Error:   "",
	}
	return c.JSON(statusCode, resp)
}

// Error 错误响应
func Error(c *echo.Context, code int, message string) error {
	resp := &Response{
		Code:    code,
		Data:    nil,
		Message: message,
		Error:   "",
	}
	return c.JSON(http.StatusOK, resp)
}

// ErrorWithStatus 错误响应（带 HTTP 状态码）
func ErrorWithStatus(c *echo.Context, httpStatus int, code int, message string) error {
	resp := &Response{
		Code:    code,
		Data:    nil,
		Message: message,
		Error:   "",
	}
	return c.JSON(httpStatus, resp)
}

// ErrorWithDetail 带详情的错误响应
func ErrorWithDetail(c *echo.Context, httpStatus int, code int, message string, detail string) error {
	resp := &Response{
		Code:    code,
		Data:    nil,
		Message: message,
		Error:   detail,
	}
	return c.JSON(httpStatus, resp)
}

// PaginatedResponse 分页响应数据
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

// SuccessPaginated 分页成功响应
//
//nolint:vet
func SuccessPaginated(c *echo.Context, items interface{}, total int64, page, pageSize int) error {
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	resp := &Response{
		Code: 0,
		Data: &PaginatedResponse{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
		Message: "success",
		Error:   "",
	}
	return c.JSON(http.StatusOK, resp)
}
