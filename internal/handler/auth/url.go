package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"project/internal/handler"
)

// Router 定义路由
type Router struct{}

func NewRouter() handler.ModuleRouter {
	return &Router{}
}

func (r *Router) RegisterRoutes(group *gin.RouterGroup, db *gorm.DB, redis *redis.Client) {
	// 创建业务处理实例
	authHandler := NewAuthHandler(db, redis)
	tokenHandler := NewTokenVerifyHandler(db, redis)

	//用户登录
	group.POST("/login", authHandler.HandlePOST)
	//验证token
	group.POST("/verify", tokenHandler.HandlePOST)
}
