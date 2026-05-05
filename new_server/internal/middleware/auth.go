package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"syc-file/config"
	"syc-file/pkg/logger"
	"syc-file/pkg/token"
	"time"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Token")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// AuthToken 全局解析token，不拦截
func AuthToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 白名单直接放行
		for _, path := range config.Conf.Whitelist {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Set("Auth", false)
				c.Next()
				return
			}
		}

		tokenStr := c.GetHeader("Token")
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			c.Set("Auth", false)
			c.Next()
			return
		}

		claims, err := token.ParseToken(tokenStr)
		if err != nil {
			logger.Logger.Warn("Token验证失败", zap.Error(err))
			c.Set("Auth", false)
			c.Next()
			return
		}

		// 检查剩余有效期，不足 refresh_expire 天则自动刷新
		remaining := time.Until(claims.ExpiresAt.Time)
		refreshThreshold := time.Duration(config.Conf.Auth.RefreshExpire) * 24 * time.Hour
		if remaining < refreshThreshold {
			newToken, err := token.GenerateToken(
				claims.UserID,
				claims.Username,
				claims.Email,
				claims.Roles,
				config.Conf.Auth.TokenExpire,
			)
			if err != nil {
				logger.Logger.Warn("Token刷新失败", zap.Error(err))
			} else {
				// 新token写回响应头，前端从 New-Token 取
				c.Header("New-Token", newToken)
				c.Header("Token-Refreshed", "true")
				logger.Logger.Info("Token已自动刷新",
					zap.Int64("user_id", claims.UserID),
					zap.String("username", claims.Username),
				)
			}
		}

		c.Set("Auth", true)
		c.Set("UserInfo", claims)
		logger.Logger.Info("Token验证成功",
			zap.Int64("user_id", claims.UserID),
			zap.String("username", claims.Username),
		)
		c.Next()
	}
}

// RequireAuth 强制登录拦截
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth, exists := c.Get("Auth")
		if !exists || auth == false {
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "未登录，请先登录",
				"data":    nil,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRole 角色校验
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInfo, exists := c.Get("UserInfo")
		if !exists {
			c.JSON(http.StatusOK, gin.H{
				"code":    401,
				"message": "未登录",
				"data":    nil,
			})
			c.Abort()
			return
		}

		claims := userInfo.(*token.Claims)
		for _, required := range roles {
			for _, role := range claims.Roles {
				if role == required {
					c.Next()
					return
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    403,
			"message": "权限不足",
			"data":    nil,
		})
		c.Abort()
	}
}
