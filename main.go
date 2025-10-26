package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"project/config"
	"project/pkg/database"
)

func main() {
	// ========== 1. 加载配置 ==========
	fmt.Println("📖 加载配置文件...")
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}
	fmt.Println("✅ 配置加载成功")

	// ========== 2. 连接PostgreSQL ==========
	fmt.Println("\n🔌 连接PostgreSQL数据库...")
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	fmt.Println("✅ PostgreSQL连接成功")

	// 测试查询
	var version string
	db.Raw("SELECT version()").Scan(&version)
	fmt.Printf("   数据库版本: %s\n", version[:50]+"...")

	// ========== 3. 连接Redis ==========
	fmt.Println("\n🔌 连接Redis...")
	rdb, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("❌ Redis连接失败: %v", err)
	}
	defer rdb.Close()
	fmt.Println("✅ Redis连接成功")

	ctx := context.Background()
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("❌ Redis Ping失败: %v", err)
	}
	fmt.Printf("   Redis响应: %s\n", pong)

	// ========== 4. 启动Web服务器 ==========
	fmt.Println("\n🚀 启动Web服务器...")
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":   "ok",
			"database": "connected",
			"redis":    "connected",
		})
	})

	// 测试路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World!",
			"status":  "success",
			"version": "1.0.0",
		})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/hello/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.JSON(200, gin.H{
			"message": fmt.Sprintf("Hello, %s!", name),
		})
	})

	// 测试数据库查询接口
	r.GET("/test/db", func(c *gin.Context) {
		var result struct {
			Now string
		}
		db.Raw("SELECT NOW() as now").Scan(&result)
		c.JSON(200, gin.H{
			"message": "数据库查询成功",
			"time":    result.Now,
		})
	})

	// 测试Redis读写接口
	r.GET("/test/redis", func(c *gin.Context) {
		err := rdb.Set(ctx, "test_key", "Hello Redis!", 0).Err()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		val, err := rdb.Get(ctx, "test_key").Result()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"message": "Redis读写成功",
			"value":   val,
		})
	})

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("\n✨ 服务器启动成功！\n")
	fmt.Printf("📍 地址: http://localhost:%d\n\n", cfg.Server.Port)
	fmt.Println("可用接口:")
	fmt.Println("  GET  /              - Hello World")
	fmt.Println("  GET  /health        - 健康检查")
	fmt.Println("  GET  /ping          - Ping测试")
	fmt.Println("  GET  /hello/:name   - 问候")
	fmt.Println("  GET  /test/db       - 测试数据库")
	fmt.Println("  GET  /test/redis    - 测试Redis")
	fmt.Println()

	if err := r.Run(addr); err != nil {
		log.Fatalf("❌ 启动服务器失败: %v", err)
	}
}
