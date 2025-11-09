package handler

import (
	"github.com/sunyuanling/server/internal/base"
	_interface "github.com/sunyuanling/server/internal/handler/user/interface"
	"github.com/sunyuanling/server/internal/handler/user/model"
	"github.com/sunyuanling/server/pkg/logger"
	"github.com/sunyuanling/server/pkg/password"
	"github.com/sunyuanling/server/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type userResetPasswordHandler struct {
	*base.BaseHandler
}

func NewUserResetPasswordHandler(db *gorm.DB, client *redis.Client) _interface.UserResetPassword {
	return &userResetPasswordHandler{
		BaseHandler: base.NewBaseHandler(db, client),
	}
}

func (h *userResetPasswordHandler) HandlePOST(c *gin.Context) {
	var req struct {
		UserName    string `json:"username" binding:"required"`
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
		Email       string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("参数错误", zap.Error(err))
		response.BadRequest(c, "传入参数错误")
		return
	}

	var user model.User
	if err := h.GetDB().Where("username = ?", req.UserName).First(&user).Error; err != nil {
		response.BadRequest(c, "用户不存在")
		return
	}

	if user.Email != req.Email {
		response.BadRequest(c, "邮箱不匹配")
		return
	}

	if !password.VerifyPassword(user.Password, req.OldPassword) {
		response.BadRequest(c, "旧密码错误")
		return
	}

	// 生成新密码
	hashed, err := password.HashPassword(req.NewPassword)
	if err != nil {
		logger.Error("密码加密失败", zap.Error(err))
		response.InternalError(c, "密码加密失败")
		return
	}

	// 更新数据库
	if err := h.GetDB().Model(&user).Update("password", hashed).Error; err != nil {
		logger.Error("密码更新失败", zap.Error(err))
		response.InternalError(c, "密码更新失败")
		return
	}

	response.Success(c, gin.H{"user": "密码重置成功"})
}
