package user

import (
	"github.com/sunyuanling/server/config"
	"github.com/sunyuanling/server/internal/handler"
	userHandler "github.com/sunyuanling/server/internal/handler/user/handler"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Router 定义路由
type Router struct {
	cfg *config.Config
}

// NewUserRouter 新建用户相关路由
func NewUserRouter(cfg *config.Config) handler.ModuleRouter {
	return &Router{
		cfg: cfg,
	}
}

// RegisterRoutes 注册用户相关路由
func (r *Router) RegisterRoutes(group *gin.RouterGroup, db *gorm.DB, client *redis.Client) {
	userRegister := userHandler.NewRegisterHandler(db, client)
	userResetPassword := userHandler.NewUserResetPasswordHandler(db, client)
	updateUserInfo := userHandler.NewUpdateUserInfoHandler(db, client, r.cfg)
	group.POST("/register", userRegister.HandlePOST)
	group.POST("/resetPassword", userResetPassword.HandlePOST)
	group.POST("/updateInfo", updateUserInfo.HandlePOST)
}
