package user

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"syc-file/internal/model"
	"syc-file/pkg/logger"
	"syc-file/pkg/password"
)

func HandlerFuncResetPassword(db *gorm.DB, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username    string `json:"username"`
			Email       string `json:"email"`
			Phone       string `json:"phone"`
			NewPassword string `json:"new_password" binding:"required"`
		}

		// 1. 绑定参数
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "参数格式错误",
				"data":    nil,
			})
			return
		}

		// 2. 校验新密码
		if len(req.NewPassword) < 6 {
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "新密码长度至少6位",
				"data":    nil,
			})
			return
		}

		// 3. 统计提供的条件数量，至少满足两个
		condCount := 0
		if req.Username != "" {
			condCount++
		}
		if req.Email != "" {
			condCount++
		}
		if req.Phone != "" {
			condCount++
		}
		if condCount < 2 {
			c.JSON(http.StatusOK, gin.H{
				"code":    400,
				"message": "请至少提供用户名、邮箱、手机号中的两项",
				"data":    nil,
			})
			return
		}

		// 4. 按条件查找用户，所有提供的条件都要匹配
		query := db.Model(&model.User{})
		if req.Username != "" {
			query = query.Where("username = ?", req.Username)
		}
		if req.Email != "" {
			query = query.Where("email = ?", req.Email)
		}
		if req.Phone != "" {
			query = query.Where("phone = ?", req.Phone)
		}

		var u model.User
		if err := query.First(&u).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{
					"code":    400,
					"message": "用户信息不匹配，请检查输入",
					"data":    nil,
				})
				return
			}
			logger.Logger.Error("查询用户失败", zap.Error(err))
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "服务器错误",
				"data":    nil,
			})
			return
		}

		// 5. 检查状态
		if u.Status == 0 {
			c.JSON(http.StatusOK, gin.H{
				"code":    403,
				"message": "账号已被禁用",
				"data":    nil,
			})
			return
		}

		// 6. 加密新密码
		hashedPassword, err := password.HashPassword(req.NewPassword)
		if err != nil {
			logger.Logger.Error("密码加密失败", zap.Error(err))
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "服务器错误",
				"data":    nil,
			})
			return
		}

		// 7. 更新密码
		if err := db.Model(&u).Update("password", hashedPassword).Error; err != nil {
			logger.Logger.Error("更新密码失败", zap.Error(err))
			c.JSON(http.StatusOK, gin.H{
				"code":    500,
				"message": "服务器错误",
				"data":    nil,
			})
			return
		}

		logger.Logger.Info("用户重置密码成功",
			zap.Uint("user_id", u.ID),
			zap.String("username", u.Username),
			zap.String("ip", c.ClientIP()),
		)

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "密码重置成功",
			"data":    nil,
		})
	}
}
