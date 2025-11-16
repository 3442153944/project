package main

import (
	"context"
	"fmt"
	"log"

	"go.uber.org/zap"

	"github.com/sunyuanling/server/config"
	"github.com/sunyuanling/server/gateway"
	"github.com/sunyuanling/server/pkg/database"
	"github.com/sunyuanling/server/pkg/logger"
	token "github.com/sunyuanling/server/pkg/tokn"
)

func main() {
	// ========== 1. 加载配置 ==========
	fmt.Println("加载配置文件...")
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	fmt.Println("配置加载成功")

	// ========== 2. 初始化日志系统 ==========
	if err := logger.Init(cfg); err != nil {
		log.Fatalf("日志初始化失败: %v", err)
	}
	defer logger.Sync() // 确保程序退出前刷新日志

	logger.Info("应用启动",
		zap.String("mode", cfg.Server.Mode),
		zap.String("version", "1.0.0"),
	)

	// 初始化token系统
	if err := token.InitGlobalTokenManager(cfg); err != nil {
		logger.Fatal("token初始化失败", zap.Error(err))
	}

	// ========== 3. 连接PostgreSQL ==========
	logger.Info("连接PostgreSQL数据库...")
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}
	logger.Info("PostgreSQL连接成功",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
	)

	// ========== 4. 连接Redis ==========
	logger.Info("连接Redis...")
	rdb, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		logger.Fatal("Redis连接失败", zap.Error(err))
	}
	defer rdb.Close()
	logger.Info("Redis连接成功",
		zap.String("host", cfg.Redis.Host),
		zap.Int("port", cfg.Redis.Port),
	)

	// 测试Redis
	ctx := context.Background()
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		logger.Fatal("Redis Ping失败", zap.Error(err))
	}
	logger.Debug("Redis Ping测试", zap.String("response", pong))

	// ========== 5. 初始化网关（传递配置） ==========
	logger.Info("初始化API网关...")
	gw := gateway.NewGateway(db, rdb, cfg)
	if gw == nil {
		logger.Fatal("网关初始化失败")
	}
	gw.SetupRoutes()

	// ========== 6. 启动服务器 ==========
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("服务器启动",
		zap.String("address", addr),
		zap.String("mode", cfg.Server.Mode),
		zap.String("file_mode", cfg.File.Mode),
		zap.Strings("allowed_paths", cfg.GetAllowedPaths()),
	)
	fmt.Printf("\n🚀 服务器运行在 http://localhost:%d\n", cfg.Server.Port)
	fmt.Printf("📁 文件存储模式: %s\n", cfg.File.Mode)
	fmt.Printf("💾 允许的存储路径: %v\n\n", cfg.GetAllowedPaths())

	if err := gw.Run(addr); err != nil {
		logger.Fatal("服务器启动失败", zap.Error(err))
	}
}
