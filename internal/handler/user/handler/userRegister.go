// internal/handler/user/handler/register.go
package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"project/internal/base"
	_interface "project/internal/handler/user/interface"
	userModel "project/internal/handler/user/model"
	"project/pkg/logger"
	"project/pkg/password" // 导入密码工具
	"project/pkg/response"
)

type registerHandler struct {
	*base.BaseHandler
}

func NewRegisterHandler(db *gorm.DB, redis *redis.Client) _interface.UserRegister {
	return &registerHandler{
		BaseHandler: base.NewBaseHandler(db, redis),
	}
}

func (h *registerHandler) HandlePOST(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Avatar   string `json:"avatar"`
		Sex      string `json:"sex"`
	}

	// 1. 绑定参数
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("注册请求参数错误", zap.Error(err))
		response.BadRequest(c, "参数格式错误")
		return
	}

	// 2. 校验密码
	if req.Password == "" {
		response.BadRequest(c, "密码不能为空")
		return
	}

	if len(req.Password) < 6 {
		response.BadRequest(c, "密码长度至少6位")
		return
	}

	// 3. 校验至少一个凭证
	if req.Username == "" && req.Email == "" && req.Phone == "" {
		response.BadRequest(c, "请提供用户名、邮箱或手机号")
		return
	}

	// 4. 校验用户名长度（匹配数据库约束）
	if req.Username != "" && len(req.Username) < 3 {
		response.BadRequest(c, "用户名长度至少3位")
		return
	}

	// 5. 检查是否已存在
	var existingUser userModel.User
	err := h.DB.Where("username = ? OR email = ? OR phone = ?",
		req.Username, req.Email, req.Phone).First(&existingUser).Error

	if err == nil {
		// 找到重复用户
		if existingUser.Username == req.Username {
			response.BadRequest(c, "用户名已存在")
			return
		}
		if existingUser.Email == req.Email && req.Email != "" {
			response.BadRequest(c, "邮箱已被注册")
			return
		}
		if existingUser.Phone == req.Phone && req.Phone != "" {
			response.BadRequest(c, "手机号已被注册")
			return
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("查询用户失败", zap.Error(err))
		response.Error(c, 500, "服务器错误")
		return
	}

	//  6. 密码加密（bcrypt）
	hashedPassword, err := password.HashPassword(req.Password)
	if err != nil {
		logger.Error("密码加密失败", zap.Error(err))
		response.Error(c, 500, "服务器错误")
		return
	}

	// 7. 创建用户
	user := userModel.User{
		Username: req.Username,
		Password: hashedPassword, //  存储加密后的密码
		Email:    req.Email,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
		Status:   1, // 默认正常状态
	}

	// 8. 插入数据库
	if err := h.DB.Create(&user).Error; err != nil {
		logger.Error("创建用户失败",
			zap.Error(err),
			zap.String("username", req.Username),
		)
		response.Error(c, 500, "注册失败")
		return
	}

	// 9. 记录日志
	logger.Info("用户注册成功",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("ip", c.ClientIP()),
	)

	// 10. 返回成功（不返回密码）
	response.Success(c, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"phone":      user.Phone,
			"avatar":     user.Avatar,
			"status":     user.Status,
			"created_at": user.CreatedAt,
		},
	})
}
