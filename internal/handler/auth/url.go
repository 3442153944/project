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
	//用户登录
	authHandler := NewAuthHandler(db, redis)
	group.POST("/login", authHandler.HandlePOST)
}
