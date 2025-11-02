package _interface

import "github.com/gin-gonic/gin"

// UserRegister 用户注册接口
type UserRegister interface {
	HandlePOST(c *gin.Context)
}
