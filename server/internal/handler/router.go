package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ModuleRouter 模块路由接口
type ModuleRouter interface {
	RegisterRoutes(group *gin.RouterGroup, db *gorm.DB, redis *redis.Client)
}
