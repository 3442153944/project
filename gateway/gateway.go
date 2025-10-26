package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"project/internal/handler"
	"project/pkg/response"
)

// Gateway 网关结构
type Gateway struct {
	router *gin.Engine
	db     *gorm.DB
	redis  *redis.Client
}

// NewGateway 创建网关
func NewGateway(db *gorm.DB, redis *redis.Client) *Gateway {
	return &Gateway{
		router: gin.Default(),
		db:     db,
		redis:  redis,
	}
}

// SetupRoutes 设置所有路由
func (g *Gateway) SetupRoutes() {
	// ========== 健康检查（不经过业务Handler）==========
	g.router.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{
			"status":   "ok",
			"database": "connected",
			"redis":    "connected",
		})
	})

	// ========== API网关 ==========
	api := g.router.Group("/api/v1")
	{
		// 测试路由组
		test := api.Group("/test")
		{
			// Redis测试 - 注册所有HTTP方法
			redisHandler := handler.NewTestRedisHandler(g.db, g.redis)
			test.GET("/redis", redisHandler.HandleGET)
			test.POST("/redis", redisHandler.HandlePOST)
			test.PUT("/redis", redisHandler.HandlePUT)       // 未重写，返回405
			test.DELETE("/redis", redisHandler.HandleDELETE) // 未重写，返回405
			test.PATCH("/redis", redisHandler.HandlePATCH)   // 未重写，返回405

			// 数据库测试
			dbHandler := handler.NewTestDBHandler(g.db, g.redis)
			test.GET("/db", dbHandler.HandleGET)
			test.POST("/db", dbHandler.HandlePOST)
			test.PUT("/db", dbHandler.HandlePUT)       // 未重写，返回405
			test.DELETE("/db", dbHandler.HandleDELETE) // 未重写，返回405
			test.PATCH("/db", dbHandler.HandlePATCH)   // 未重写，返回405
		}

		// 认证路由组（后续添加）
		// auth := api.Group("/auth")
		// {
		//     authHandler := handler.NewAuthHandler(g.db, g.redis)
		//     auth.POST("/login", authHandler.HandlePOST)
		//     auth.POST("/register", authHandler.HandlePOST)
		// }
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

// GetRouter 获取路由引擎
func (g *Gateway) GetRouter() *gin.Engine {
	return g.router
}

// Run 启动网关
func (g *Gateway) Run(addr string) error {
	return g.router.Run(addr)
}
