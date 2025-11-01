package auth

import (
	"context"
	"errors"
	"fmt"
	token "project/pkg/tokn"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"project/internal/base"
	model "project/internal/handler/auth/modle"
	"project/pkg/logger"
	"project/pkg/response"
)

type Handler struct {
	*base.BaseHandler
}

func NewAuthHandler(db *gorm.DB, redis *redis.Client) *Handler {
	return &Handler{
		BaseHandler: base.NewBaseHandler(db, redis),
	}
}

// HandlePOST Login 登录
func (h *Handler) HandlePOST(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("登录请求参数错误",
			zap.String("error", err.Error()),
			zap.String("ip", c.ClientIP()),
		)
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 校验：至少提供一个登录凭证
	if req.Username == "" && req.Email == "" && req.Phone == "" {
		logger.Warn("登录凭证为空", zap.String("ip", c.ClientIP()))
		response.BadRequest(c, "请提供用户名、邮箱或手机号")
		return
	}

	// 登录标识（用于Redis key）
	loginKey := req.Username
	if loginKey == "" {
		loginKey = req.Email
	}
	if loginKey == "" {
		loginKey = req.Phone
	}

	// 检查登录失败次数
	redisKey := fmt.Sprintf("login_fail:%s", loginKey)
	ctx := context.Background()

	failCount, err := h.Redis.Get(ctx, redisKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		logger.Error("Redis查询失败",
			zap.String("key", redisKey),
			zap.Error(err),
		)
	}

	if failCount >= 5 {
		logger.Warn("登录失败次数超限",
			zap.String("login_key", loginKey),
			zap.Int("fail_count", failCount),
			zap.String("ip", c.ClientIP()),
		)
		response.TooManyRequests(c, "登录失败次数过多，请5分钟后再试")
		return
	}

	// 查询用户
	var user model.User
	err = h.DB.Where("username = ? OR email = ? OR phone = ?",
		req.Username, req.Email, req.Phone).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("用户不存在",
				zap.String("login_key", loginKey),
				zap.String("ip", c.ClientIP()),
			)
		} else {
			logger.Error("查询用户失败",
				zap.String("login_key", loginKey),
				zap.Error(err),
			)
		}

		// 记录失败次数
		h.incrLoginFail(redisKey)

		response.Unauthorized(c, "用户名或密码错误")
		return
	}

	// 验证密码
	if user.Password != req.Password {
		logger.Warn("密码错误",
			zap.Uint("user_id", user.ID),
			zap.String("username", user.Username),
			zap.String("ip", c.ClientIP()),
		)
		println("密码错误")
		// 记录失败次数
		h.incrLoginFail(redisKey)
		println("尝试记录失败次数")

		response.Unauthorized(c, "用户名或密码错误")
		return
	}

	// 登录成功：清除失败记录
	h.Redis.Del(ctx, redisKey)

	logger.Info("用户登录成功",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("ip", c.ClientIP()),
	)
	//生成token

	generateToken, err := h.generateToken(&user)

	if err != nil {
		logger.Error("生成token失败", zap.Error(err))
		response.Error(c, 500, "服务器内部错误")
		return
	}

	response.Success(c, gin.H{"generateToken": generateToken, "user": user})
}

// 在 incrLoginFail 函数里加日志排查
func (h *Handler) incrLoginFail(redisKey string) {
	ctx := context.Background()

	// 增加计数
	result := h.Redis.Incr(ctx, redisKey)
	if result.Err() != nil {
		logger.Error("Redis Incr失败", zap.Error(result.Err()))
		return
	}

	// 设置过期时间：5分钟
	expireResult := h.Redis.Expire(ctx, redisKey, 5*time.Minute)
	if expireResult.Err() != nil {
		logger.Error("Redis Expire失败", zap.Error(expireResult.Err()))
		return
	}

	logger.Info("设置过期时间成功", zap.Duration("ttl", 5*time.Minute))
}

func (h *Handler) generateToken(user *model.User) (string, error) {
	payload := map[string]interface{}{
		"user_id":   user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"role":      user.Role,
		"status":    user.Status,
		"phone":     user.Phone,
		"lastLogin": user.LastLogin,
		"createAt":  user.CreatedAt,
		"updateAt":  user.UpdatedAt,
	}
	resToken, err := token.GetGlobalTokenManager().GenerateTokenFromMap(payload)
	return resToken, err
}
