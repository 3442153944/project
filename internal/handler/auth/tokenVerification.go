package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"project/internal/base"
	_interface "project/internal/handler/auth/interface"
	"project/pkg/response"
)

// TokenHandler Handler 继承基类
type tokenHandler struct {
	*base.BaseHandler
}

func NewTokenVerifyHandler(db *gorm.DB, redis *redis.Client) _interface.TokenVerifyHandler {
	return &tokenHandler{
		BaseHandler: base.NewBaseHandler(db, redis),
	}
}

func (h *tokenHandler) HandlePOST(c *gin.Context) {
	//  获取认证状态（使用GetBool更简洁）
	isAuth := c.GetBool("Auth")

	// 未认证直接返回
	if !isAuth {
		response.Unauthorized(c, "token验证失败")
		return
	}

	//  已认证，返回用户信息
	userInfo, _ := c.Get("UserInfo")
	response.Success(c, gin.H{
		"msg":      "token有效",
		"userInfo": userInfo,
	})
}
