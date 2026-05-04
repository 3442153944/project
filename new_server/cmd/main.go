package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"syc-file/config"
	"syc-file/internal/database"
	"syc-file/internal/handler"
	"syc-file/pkg/logger"
)

func main() {
	// 1. 初始化配置 (Viper)
	if err := config.Init(); err != nil {
		panic("配置初始化失败: " + err.Error())
	}

	// 2. 初始化日志 (Zap + Lumberjack)
	// 将 config 模块中解析好的 Log 配置传给 logger 模块
	if err := logger.Init(config.Conf.Log); err != nil {
		panic("日志初始化失败: " + err.Error())
	}
	// 程序退出前刷新日志缓冲
	defer func(Logger *zap.Logger) {
		err := Logger.Sync()
		if err != nil {
			logger.Logger.Error("日志缓冲刷新失败", zap.Error(err))
		}
	}(logger.Logger)

	logger.Logger.Info("配置与日志初始化成功", zap.Int("port", config.Conf.Server.Port))

	// 3. 初始化 Gin 引擎
	// 以前是 r := gin.Default()，现在改为 gin.New()，并手动挂载我们的 Zap 中间件和默认的恢复中间件
	r := gin.New()
	//r.Use(middleware.ZapLogger(), gin.Recovery())

	//建立数据库连接
	db, err := database.InitMySQL(config.Conf.DB)
	if err != nil {
		logger.Logger.Error("数据库连接失败", zap.Error(err))
	}

	// 4. 注册路由
	r.GET("/ping", func(c *gin.Context) {
		// 在业务代码里打印日志的正确姿势
		logger.Logger.Info("收到 ping 请求")
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	handler.RegisterRouters(r, db)

	// 5. 启动服务
	addr := fmt.Sprintf(":%d", config.Conf.Server.Port)
	logger.Logger.Info("服务器准备启动", zap.String("addr", addr))

	if err := r.Run(addr); err != nil {
		logger.Logger.Fatal("服务器启动失败", zap.Error(err))
	}
}
