package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/sunyuanling/server/internal/base"
	_interface "github.com/sunyuanling/server/internal/handler/auth/interface"
	model "github.com/sunyuanling/server/internal/handler/auth/modle"
	"github.com/sunyuanling/server/pkg/logger"
	"github.com/sunyuanling/server/pkg/response"
	tokenPkg "github.com/sunyuanling/server/pkg/tokn"
)

type tokenHandler struct {
	*base.BaseHandler
}

func NewTokenVerifyHandler(db *gorm.DB, redis *redis.Client) _interface.TokenVerifyHandler {
	return &tokenHandler{
		BaseHandler: base.NewBaseHandler(db, redis),
	}
}

func (h *tokenHandler) HandlePOST(c *gin.Context) {
	isAuth := c.GetBool("Auth")
	if !isAuth {
		response.Unauthorized(c, "token验证失败")
		return
	}

	payloadRaw, exists := c.Get("UserInfo")
	if !exists || payloadRaw == nil {
		response.InternalError(c, "用户信息获取失败")
		return
	}

	payload, ok := payloadRaw.(*tokenPkg.TokenPayload)
	if !ok {
		response.InternalError(c, "用户信息类型错误")
		return
	}

	// 查库获取完整用户信息
	var user model.User
	if err := h.DB.First(&user, payload.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("verify: 用户不存在", zap.Int64("user_id", payload.UserID))
			response.Unauthorized(c, "用户不存在")
		} else {
			logger.Error("verify: 查询用户失败", zap.Error(err))
			response.InternalError(c, "查询用户失败")
		}
		return
	}

	response.Success(c, gin.H{
		"msg": "token有效",
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"phone":      user.Phone,
			"avatar":     user.Avatar,
			"role":       user.Role,
			"status":     user.Status,
			"last_login": user.LastLogin,
			"created_at": user.CreatedAt,
		},
	})
}
