package files

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"project/internal/handler"
	filesHandler "project/internal/handler/files/handler"
)

type Router struct {
}

func NewRouter() handler.ModuleRouter {
	return &Router{}
}

// RegisterRoutes 修正参数顺序：group 在前，db 和 redis 在后
func (r *Router) RegisterRoutes(group *gin.RouterGroup, db *gorm.DB, redis *redis.Client) {
	// 创建处理器实例
	getAvailableDiskList := filesHandler.NewGetAvailableDiskList(db, redis)
	traverseDirectory := filesHandler.NewTraverseDirectory(db, redis)

	// 注册路由
	group.POST("/available-disks", getAvailableDiskList.HandlerPOST)
	group.POST("/traverse-directory", traverseDirectory.HandlerPOST)
}
