package main

import (
	"context"
	"fmt"
	"log"
	"project/config"
	"project/gateway"
	"project/pkg/database"
	"project/pkg/logger"

	"go.uber.org/zap"
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

	// ========== 5. 初始化网关 ==========
	logger.Info("初始化API网关...")
	gw := gateway.NewGateway(db, rdb)
	gw.SetupRoutes()

	// ========== 6. 启动服务器 ==========
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("服务器启动",
		zap.String("address", addr),
		zap.String("mode", cfg.Server.Mode),
	)
	fmt.Printf("\n🚀 服务器运行在 http://localhost:%d\n\n", cfg.Server.Port)

	if err := gw.Run(addr); err != nil {
		logger.Fatal("服务器启动失败", zap.Error(err))
	}
}
