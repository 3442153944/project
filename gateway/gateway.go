package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"project/internal/handler/test"
	// "project/internal/handler/auth"  // 未来添加
	"project/pkg/response"
)

type Gateway struct {
	router *gin.Engine
	db     *gorm.DB
	redis  *redis.Client
}

func NewGateway(db *gorm.DB, redis *redis.Client) *Gateway {
	return &Gateway{
		router: gin.Default(),
		db:     db,
		redis:  redis,
	}
}

func (g *Gateway) SetupRoutes() {
	// ========== 健康检查 ==========
	g.router.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{
			"status":   "ok",
			"database": "connected",
			"redis":    "connected",
		})
	})

	// ========== API网关 - 只注册模块入口 ==========
	api := g.router.Group("/api")
	{
		// 注册test模块路由
		testGroup := api.Group("/test")
		test.NewRouter().RegisterRoutes(testGroup, g.db, g.redis)

		// 注册auth模块路由（未来添加）
		// authGroup := api.Group("/auth")
		// auth.NewRouter().RegisterRoutes(authGroup, g.db, g.redis)
	}

	// ========== 根路径 ==========
	g.router.GET("/", func(c *gin.Context) {
		response.Success(c, gin.H{
			"message": "Welcome to File Sync API Gateway",
			"version": "1.0.0",
		})
	})

	// ========== 404处理 ==========
	g.router.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "接口不存在")
	})
}

func (g *Gateway) GetRouter() *gin.Engine {
	return g.router
}

func (g *Gateway) Run(addr string) error {
	return g.router.Run(addr)
}
