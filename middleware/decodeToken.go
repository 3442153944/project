package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
		if token == "" {
			c.Set("Auth", false)
			c.Next()
			return
		}

		payload, err := tokenFunc.GetGlobalTokenManager().ValidateToken(token)
		if err != nil {
			logger.Warn("Token验证失败: " + err.Error())
			c.Set("Auth", false)
			c.Next()
			return
		}

		// payload 是 *TokenPayload
		if payload == nil {
			logger.Warn("Token验证成功，但 payload 为空")
			c.Set("Auth", false)
			c.Next()
			return
		}

		// 设置到上下文
		c.Set("Auth", true)
		c.Set("UserInfo", payload)

		logger.Info("Token验证成功, UserInfo:", zap.Any("user_info", payload))
		c.Next()
	}
}

func interfaceMapToStringMap(m interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	switch mm := m.(type) {
	case map[interface{}]interface{}:
		for k, v := range mm {
			res[fmt.Sprintf("%v", k)] = v
		}
	case map[string]interface{}:
		res = mm
	default:
		return nil
	}
	return res
}
