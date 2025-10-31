package middleware

import (
	"github.com/gin-gonic/gin"
	"project/pkg/logger"
)

// 中间件跨域处理
func CORS() gin.HandlerFunc {
	logger.Info("CORS函数初始化")
	return func(c *gin.Context) {
		logger.Info("CORS函数执行")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Next()
	}
}

// 中间件token认证处理
func AuthToken() gin.HandlerFunc {
	logger.Info("AuthToken函数初始化")
	return func(context *gin.Context) {
		var token = context.GetHeader("Token")
		if token == "" {
			logger.Warn("token为空")
			context.Set("Auth", false)
		} else {
			println("token:", token)
			context.Set("Auth", true)
		}
		context.Next()
	}
}
