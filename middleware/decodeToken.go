package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"project/pkg/logger"
	tokenFunc "project/pkg/tokn"
)

// CORS 中间件跨域处理
func CORS() gin.HandlerFunc {
	logger.Info("CORS中间件已加载")
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Token") //  加上Token
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		//  OPTIONS请求直接返回，不执行后续逻辑
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// AuthToken 中间件token认证处理（全局，不拦截）
func AuthToken() gin.HandlerFunc {
	logger.Info("AuthToken中间件已加载")
	return func(c *gin.Context) {
		token := c.GetHeader("Token")

		//  Token为空，标记为未认证
		if token == "" {
			c.Set("Auth", false)
			c.Next() //  只调用一次
			return
		}

		//  验证Token
		userInfo, err := tokenFunc.GetGlobalTokenManager().ValidateToken(token)
		if err != nil {
			logger.Warn("Token验证失败: " + err.Error())
			c.Set("Auth", false)
			c.Next() //  只调用一次
			return
		}

		//  验证成功
		c.Set("Auth", true)
		c.Set("UserInfo", userInfo)
		c.Next() //  只调用一次
	}
}
