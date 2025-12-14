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

// GetFile 获取文件
type GetFile interface {
	HandlerGET(c *gin.Context)
}

// FileMsg 文件传参实时信息
type FileMsg interface {
	HandlerPOST(c *gin.Context)
}

// FileUpload 文件上传
type FileUpload interface {
	HandlerPOST(c *gin.Context)
}
