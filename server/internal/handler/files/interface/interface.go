package _interface

import "github.com/gin-gonic/gin"

// GetAvailableDiskList 获取可用磁盘列表
type GetAvailableDiskList interface {
	HandlerPOST(c *gin.Context)
}

// TraverseDirectory 遍历目录
type TraverseDirectory interface {
	HandlerPOST(c *gin.Context)
}
