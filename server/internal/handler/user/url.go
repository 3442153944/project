package user

import (
	"github.com/sunyuanling/server/internal/handler"
	userHandler "github.com/sunyuanling/server/internal/handler/user/handler"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Router 定义路由
type Router struct{}

// NewUserRouter 新建用户相关路由
func NewUserRouter() handler.ModuleRouter {
	return &Router{}
}

// RegisterRoutes 注册用户相关路由
func (r *Router) RegisterRoutes(group *gin.RouterGroup, db *gorm.DB, client *redis.Client) {
	userRegister := userHandler.NewRegisterHandler(db, client)
	userResetPassword := userHandler.NewUserResetPasswordHandler(db, client)
	group.POST("/register", userRegister.HandlePOST)
	group.POST("/resetPassword", userResetPassword.HandlePOST)
}
