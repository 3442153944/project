package response

import "github.com/gin-gonic/gin"

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMsg 成功响应（自定义消息）
func SuccessWithMsg(c *gin.Context, message string, data interface{}) {
	c.JSON(200, Response{
		Code:    200,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequest 400错误
func BadRequest(c *gin.Context, message string) {
	Error(c, 400, message)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, message string) {
	Error(c, 401, message)
}

// Forbidden 403错误
func Forbidden(c *gin.Context, message string) {
	Error(c, 403, message)
}

// NotFound 404错误
func NotFound(c *gin.Context, message string) {
	Error(c, 404, message)
}

// MethodNotAllowed 405错误（方法未实现/不允许）
func MethodNotAllowed(c *gin.Context, message string) {
	c.Header("Allow", "GET, POST, PUT, DELETE, PATCH") // 可选：告知允许的方法
	Error(c, 405, message)
}

// InternalError 500错误
func InternalError(c *gin.Context, message string) {
	Error(c, 500, message)
}

// TooManyRequests 429错误
func TooManyRequests(c *gin.Context, msg string) {
	c.JSON(429, gin.H{
		"code": 429,
		"msg":  msg,
	})
}
