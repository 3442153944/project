package _interface

import "github.com/gin-gonic/gin"

// UserRegister 用户注册接口
type UserRegister interface {
	HandlePOST(c *gin.Context)
}

// UserResetPassword 用户重置密码接口
type UserResetPassword interface {
	HandlePOST(c *gin.Context)
}
