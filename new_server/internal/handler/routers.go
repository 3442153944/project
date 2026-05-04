package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func RegisterRouters(r *gin.Engine, db *gorm.DB) {
	v1 := r.Group("/v1")
	public := v1.Group("")
	{
		public.POST("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})
	}
}
