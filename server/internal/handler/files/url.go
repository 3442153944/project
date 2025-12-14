package files

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/sunyuanling/server/config"
	"github.com/sunyuanling/server/internal/handler"
	filesHandler "github.com/sunyuanling/server/internal/handler/files/handler"
)

type Router struct {
	cfg *config.Config
}

// NewRouter 创建路由（需要传入配置）
func NewRouter(cfg *config.Config) handler.ModuleRouter {
	return &Router{
		cfg: cfg,
	}
}

// RegisterRoutes 注册路由
func (r *Router) RegisterRoutes(group *gin.RouterGroup, db *gorm.DB, redis *redis.Client) {
	// 创建处理器实例（传递配置）
	getAvailableDiskList := filesHandler.NewGetAvailableDiskList(db, redis, r.cfg)
	traverseDirectory := filesHandler.NewTraverseDirectory(db, redis)
	getFile := filesHandler.NewGetFile(db, redis, r.cfg)

	// 注册路由
	group.POST("/available-disks", getAvailableDiskList.HandlerPOST)
	group.POST("/traverse-directory", traverseDirectory.HandlerPOST)
	group.GET("/get-file", getFile.HandlerGET)
}
