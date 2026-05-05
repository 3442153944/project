package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"net/http"
	"syc-file/internal/handler/user"
	"syc-file/internal/middleware"
)

func RegisterRouters(r *gin.Engine, db *gorm.DB, redisClient *redis.Client) {
	v1 := r.Group("/v1")
	v1.Use(middleware.AuthToken())
	public := v1.Group("")
	{
		public.POST("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})
	}
	user.RegisterUserRouter(v1, db, redisClient)
	//接入token中间件
	// 需要登录的路由
	private := v1.Group("")
	private.Use(middleware.RequireAuth())
	{
		// 后续需要登录的路由都注册在这里
		// file.RegisterFileRouter(private, db, redisClient)
	}
}
