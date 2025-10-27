package test

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"project/internal/handler"
)

// Router test模块路由
type Router struct{}

// NewRouter 创建test路由
func NewRouter() handler.ModuleRouter {
	return &Router{}
}

// RegisterRoutes 注册test模块的所有路由
func (r *Router) RegisterRoutes(group *gin.RouterGroup, db *gorm.DB, redis *redis.Client) {
	// Redis测试路由
	redisHandler := NewTestRedisHandler(db, redis)
	group.GET("/redis", redisHandler.HandleGET)
	group.POST("/redis", redisHandler.HandlePOST)
	group.PUT("/redis", redisHandler.HandlePUT)
	group.DELETE("/redis", redisHandler.HandleDELETE)
	group.PATCH("/redis", redisHandler.HandlePATCH)

	// 数据库测试路由
	dbHandler := NewTestDBHandler(db, redis)
	group.GET("/db", dbHandler.HandleGET)
	group.POST("/db", dbHandler.HandlePOST)
	group.PUT("/db", dbHandler.HandlePUT)
	group.DELETE("/db", dbHandler.HandleDELETE)
	group.PATCH("/db", dbHandler.HandlePATCH)
	//默认路由，返回404

}
