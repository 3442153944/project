package _interface

import "github.com/gin-gonic/gin"

// AuthHandler 认证接口或者说登录接口
type AuthHandler interface {
	HandlePOST(c *gin.Context)
}

// TokenVerifyHandler token有效性验证接口
type TokenVerifyHandler interface {
	HandlePOST(c *gin.Context)
}
