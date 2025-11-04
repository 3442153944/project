package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"project/internal/handler/auth"
	"project/internal/handler/files"
	"project/internal/handler/test"
	"project/internal/handler/user"
	"project/middleware"
	"project/pkg/logger"
	"project/pkg/response"
	"project/websocket"
)

type Gateway struct {
	router    *gin.Engine
	db        *gorm.DB
	redis     *redis.Client
	wsHandler *websocket.Handler
}

func NewGateway(db *gorm.DB, redis *redis.Client) *Gateway {
	router := gin.Default()
	err := router.SetTrustedProxies([]string{"127.0.0.1", "192.168.0.0/16", "::1"})
	if err != nil {
		logger.Error("SetTrustedProxies失败", zap.Error(err))
		return nil
	}
	router.Use(middleware.CORS())
	router.Use(middleware.AuthToken())

	// 初始化WebSocket
	wsHandler := websocket.InitWebSocket(db)

	return &Gateway{
		router:    router,
		db:        db,
		redis:     redis,
		wsHandler: wsHandler,
	}
}

func (g *Gateway) SetupRoutes() {
	// ========== 健康检查 ==========
	g.router.GET("/health", func(c *gin.Context) {
		onlineCount := websocket.GetOnlineUserCount()
		response.Success(c, gin.H{
			"status":       "ok",
			"database":     "connected",
			"redis":        "connected",
			"websocket":    "active",
			"online_users": onlineCount,
		})
	})

	// ========== API网关 - 只注册模块入口 ==========
	api := g.router.Group("/api")
	{
		// 注册test模块路由
		testGroup := api.Group("/test")
		test.NewRouter().RegisterRoutes(testGroup, g.db, g.redis)

		// 注册auth模块路由
		authGroup := api.Group("/auth")
		auth.NewRouter().RegisterRoutes(authGroup, g.db, g.redis)

		// 注册user模块
		userGroup := api.Group("/user")
		user.NewUserRouter().RegisterRoutes(userGroup, g.db, g.redis)

		// 注册WebSocket路由
		g.wsHandler.RegisterRoutes(api)

		//注册files模块
		filesGroup := api.Group("/files")
		files.NewRouter().RegisterRoutes(filesGroup, g.db, g.redis)
	}

	// ========== 根路径 ==========
	g.router.GET("/", func(c *gin.Context) {
		response.Success(c, gin.H{
			"message":      "Welcome to File Sync API Gateway",
			"version":      "1.0.0",
			"websocket":    "ws://localhost:9999/api/ws/connect",
			"online_users": websocket.GetOnlineUserCount(),
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
	// 启动时输出WebSocket信息
	logger.Info("WebSocket服务已启用",
		zap.String("endpoint", "/api/ws/connect"),
	)
	return g.router.Run(addr)
}

// Shutdown 优雅关闭
func (g *Gateway) Shutdown() {
	logger.Info("正在关闭网关服务...")

	// 关闭WebSocket连接
	if hub := websocket.GetHub(); hub != nil {
		hub.Shutdown()
	}

	logger.Info("网关服务已关闭")
}
