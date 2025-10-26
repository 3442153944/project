package main

import (
	"context"
	"fmt"
	"log"
	"project/config"
	"project/gateway"
	"project/pkg/database"
)

func main() {
	// ========== 1. 加载配置 ==========
	fmt.Println("加载配置文件...")
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	fmt.Println("配置加载成功")

	// ========== 2. 连接PostgreSQL ==========
	fmt.Println("\n 连接PostgreSQL数据库...")
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("PostgreSQL连接成功")

	// ========== 3. 连接Redis ==========
	fmt.Println("\n连接Redis...")
	rdb, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Redis连接失败: %v", err)
	}
	defer rdb.Close()
	fmt.Println("Redis连接成功")

	// 测试Redis
	ctx := context.Background()
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Redis Ping失败: %v", err)
	}
	fmt.Printf("   Redis响应: %s\n", pong)

	// ========== 4. 初始化网关 ==========
	fmt.Println("\n初始化API网关...")
	gw := gateway.NewGateway(db, rdb)
	gw.SetupRoutes()

	// ========== 5. 启动服务器 ==========
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("\n 服务器启动成功！\n")
	fmt.Printf("地址: http://localhost:%d\n\n", cfg.Server.Port)

	if err := gw.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
