package middleware

import (
	"fmt"
	"github.com/sunyuanling/server/pkg/logger"
	tokenFunc "github.com/sunyuanling/server/pkg/tokn"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CORS 中间件跨域处理
func CORS() gin.HandlerFunc {
	logger.Info("CORS中间件已加载")
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Token")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// AuthToken 中间件token认证处理（全局，不拦截）
// 支持从 Header 或 Query 参数获取 token
func AuthToken() gin.HandlerFunc {
	logger.Info("AuthToken中间件已加载")
	return func(c *gin.Context) {
		// 1. 优先从 Header 获取 token
		token := c.GetHeader("Token")

		// 2. 如果 Header 没有，尝试从 Query 参数获取
		if token == "" {
			token = c.Query("token")
			if token != "" {
				logger.Info("从 Query 参数获取到 token")
			}
		} else {
			logger.Info("从 Header 获取到 token")
		}
		logger.Info("token: " + token)

		// 3. 如果都没有 token，标记为未认证
		if token == "" {
			c.Set("Auth", false)
			c.Next()
			return
		}

		// 4. 验证 token
		payload, err := tokenFunc.GetGlobalTokenManager().ValidateToken(token)
		if err != nil {
			logger.Warn("Token验证失败: " + err.Error())
			c.Set("Auth", false)
			c.Next()
			return
		}

		// 5. 检查 payload
		if payload == nil {
			logger.Warn("Token验证成功，但 payload 为空")
			c.Set("Auth", false)
			c.Next()
			return
		}

		// 6. 设置到上下文
		c.Set("Auth", true)
		c.Set("UserInfo", payload)

		logger.Info("Token验证成功",
			zap.Uint("user_id", uint(payload.UserID)),
			zap.String("username", payload.Username),
		)
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
