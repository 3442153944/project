// handler/FileUpload.go
package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sunyuanling/server/config"
	"github.com/sunyuanling/server/internal/base"
	_interface "github.com/sunyuanling/server/internal/handler/files/interface"
	"github.com/sunyuanling/server/internal/model"
	"github.com/sunyuanling/server/pkg/logger"
	"github.com/sunyuanling/server/pkg/response"
	tokenFunc "github.com/sunyuanling/server/pkg/tokn"
	"github.com/sunyuanling/server/websocket"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type fileUpload struct {
	*base.BaseHandler
	cfg *config.Config
}

// FileUploadRequest 上传请求参数
type FileUploadRequest struct {
	Path   string `json:"path" binding:"required"`   // 目标目录路径
	Name   string `json:"name" binding:"required"`   // 文件名
	Action string `json:"action" binding:"required"` // 操作类型: upload / check
}

// HandlerPOST 处理文件上传请求（POST 方法）
func (f *fileUpload) HandlerPOST(c *gin.Context) {
	// 1. 验证登录状态
	isAuth := c.GetBool("Auth")
	if !isAuth {
		response.Unauthorized(c, "请先登录")
		return
	}

	// 2. 获取用户信息
	userInfo, exists := c.Get("UserInfo")
	if !exists || userInfo == nil {
		response.InternalError(c, "用户信息获取失败")
		return
	}

	payload, ok := userInfo.(*tokenFunc.TokenPayload)
	if !ok {
		response.InternalError(c, "用户信息类型错误")
		return
	}
	userID := uint(payload.UserID)

	// 3. 解析请求参数
	var req FileUploadRequest

	// 根据 Content-Type 决定如何解析
	contentType := c.GetHeader("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// multipart 请求，从 form 获取参数
		req.Path = c.PostForm("path")
		req.Name = c.PostForm("name")
		req.Action = c.PostForm("action")
	} else {
		// JSON 请求
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "参数解析失败: "+err.Error())
			return
		}
	}

	// 4. 验证参数
	if req.Path == "" || req.Name == "" {
		response.BadRequest(c, "缺少必要参数 path 或 name")
		return
	}

	if req.Action != "check" && req.Action != "upload" {
		response.BadRequest(c, "无效的 action 参数（可选: check / upload）")
		return
	}

	logger.Info("收到上传请求",
		zap.String("path", req.Path),
		zap.String("name", req.Name),
		zap.String("action", req.Action),
		zap.Uint("user_id", userID),
	)

	// 5. 标准化路径（支持正斜杠格式）
	normalizedPath := filepath.FromSlash(req.Path)

	// 6. 智能构建完整路径
	var fullPath string
	if filepath.Base(normalizedPath) == req.Name {
		fullPath = normalizedPath
		logger.Info("使用完整路径",
			zap.String("path", req.Path),
			zap.String("name", req.Name),
		)
	} else {
		fullPath = filepath.Join(normalizedPath, req.Name)
		logger.Info("拼接路径",
			zap.String("dir", req.Path),
			zap.String("name", req.Name),
			zap.String("full_path", fullPath),
		)
	}

	// 7. 验证路径安全性
	if !f.isPathAllowed(fullPath) {
		logger.Warn("上传路径访问被拒绝",
			zap.Uint("user_id", userID),
			zap.String("path", fullPath),
		)
		response.Forbidden(c, "无权访问该路径")
		return
	}

	// 8. 验证文件扩展名
	ext := filepath.Ext(req.Name)
	if !f.cfg.IsExtensionAllowed(ext) {
		logger.Warn("文件扩展名不允许",
			zap.Uint("user_id", userID),
			zap.String("file_name", req.Name),
			zap.String("ext", ext),
		)
		response.BadRequest(c, "不允许上传该类型的文件")
		return
	}

	// 9. 验证文件名长度
	if len(req.Name) > f.cfg.File.Upload.MaxFilenameLength {
		logger.Warn("文件名过长",
			zap.Uint("user_id", userID),
			zap.String("file_name", req.Name),
			zap.Int("length", len(req.Name)),
		)
		response.BadRequest(c, fmt.Sprintf("文件名过长（最大 %d 字符）", f.cfg.File.Upload.MaxFilenameLength))
		return
	}

	// 10. 根据 action 类型处理
	switch req.Action {
	case "check":
		f.handleCheck(c, fullPath, req.Name, userID)
	case "upload":
		f.handleUpload(c, fullPath, req.Name, userID)
	}
}

// handleCheck 检查文件是否存在
func (f *fileUpload) handleCheck(c *gin.Context, fullPath, fileName string, userID uint) {
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，可以上传
			response.Success(c, gin.H{
				"exists":     false,
				"can_upload": true,
				"file_name":  fileName,
				"path":       fullPath,
			})
		} else {
			logger.Error("检查文件状态失败",
				zap.Error(err),
				zap.String("path", fullPath),
			)
			response.InternalError(c, "检查文件状态失败")
		}
		return
	}

	// 文件已存在
	response.Success(c, gin.H{
		"exists":      true,
		"can_upload":  false,
		"file_name":   fileName,
		"file_size":   fileInfo.Size(),
		"path":        fullPath,
		"modified_at": fileInfo.ModTime(),
	})
}

// handleUpload 处理上传逻辑
func (f *fileUpload) handleUpload(c *gin.Context, fullPath, fileName string, userID uint) {
	// 1. 检查文件是否已存在
	if _, err := os.Stat(fullPath); err == nil {
		logger.Warn("文件已存在",
			zap.Uint("user_id", userID),
			zap.String("path", fullPath),
		)
		response.BadRequest(c, "文件已存在，请先删除或重命名")
		return
	}

	// 2. 获取上传的文件
	fileHeader, err := c.FormFile("file")
	if err != nil {
		logger.Error("获取上传文件失败",
			zap.Error(err),
			zap.Uint("user_id", userID),
		)
		response.BadRequest(c, "请选择要上传的文件")
		return
	}

	// 3. 验证文件大小
	if fileHeader.Size > f.cfg.File.Upload.MaxFileSize {
		logger.Warn("文件大小超过限制",
			zap.Uint("user_id", userID),
			zap.String("file_name", fileName),
			zap.Int64("file_size", fileHeader.Size),
			zap.Int64("max_size", f.cfg.File.Upload.MaxFileSize),
		)
		response.BadRequest(c, fmt.Sprintf("文件大小超过限制（最大 %d MB）", f.cfg.File.Upload.MaxFileSize/1024/1024))
		return
	}

	// 4. 确保目标目录存在
	targetDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		logger.Error("创建目标目录失败",
			zap.Error(err),
			zap.String("path", targetDir),
		)
		response.InternalError(c, "创建目标目录失败")
		return
	}

	// 5. 创建上传历史记录
	history := &model.UploadHistory{
		UserID:       userID,
		FileName:     fileName,
		OriginalName: fileHeader.Filename,
		FileSize:     fileHeader.Size,
		FileType:     fileHeader.Header.Get("Content-Type"),
		StoragePath:  fullPath,
		UploadStatus: model.UploadStatusPending,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.GetHeader("User-Agent"),
	}

	if err := f.DB.Create(history).Error; err != nil {
		logger.Error("创建上传历史失败",
			zap.Error(err),
			zap.Uint("user_id", userID),
		)
	}

	// 6. 标记为上传中
	if history.ID > 0 {
		_ = history.MarkAsUploading(f.DB)
	}

	// 7. 发送 WebSocket 通知：开始上传
	_ = websocket.SendToUser(userID, "file_upload", map[string]interface{}{
		"event":        "start",
		"file_name":    fileName,
		"file_size":    fileHeader.Size,
		"history_id":   history.ID,
		"storage_path": fullPath,
	})

	// 8. 保存文件
	if err := c.SaveUploadedFile(fileHeader, fullPath); err != nil {
		logger.Error("保存文件失败",
			zap.Error(err),
			zap.String("path", fullPath),
		)
		if history.ID > 0 {
			_ = history.MarkAsFailed(f.DB, "保存文件失败: "+err.Error())
		}
		response.InternalError(c, "保存文件失败")
		return
	}

	// 9. 标记为完成
	if history.ID > 0 {
		_ = history.MarkAsCompleted(f.DB)
	}

	// 10. 发送 WebSocket 通知：上传完成
	_ = websocket.SendToUser(userID, "file_upload", map[string]interface{}{
		"event":        "completed",
		"file_name":    fileName,
		"file_size":    fileHeader.Size,
		"history_id":   history.ID,
		"storage_path": fullPath,
	})

	logger.Info("文件上传完成",
		zap.Uint("user_id", userID),
		zap.String("file_name", fileName),
		zap.Int64("file_size", fileHeader.Size),
		zap.String("path", fullPath),
	)

	// 11. 返回成功响应
	response.Success(c, gin.H{
		"history_id":    history.ID,
		"file_name":     fileName,
		"original_name": fileHeader.Filename,
		"file_size":     fileHeader.Size,
		"storage_path":  fullPath,
	})
}

// isPathAllowed 路径安全检查
func (f *fileUpload) isPathAllowed(path string) bool {
	cleanPath := filepath.Clean(path)
	cleanPathUpper := strings.ToUpper(cleanPath)

	if !filepath.IsAbs(cleanPath) {
		logger.Warn("拒绝相对路径",
			zap.String("path", path),
			zap.String("cleaned", cleanPath),
		)
		return false
	}

	if strings.Contains(path, "..") {
		logger.Warn("拒绝包含..的路径", zap.String("path", path))
		return false
	}

	allowedPaths := f.cfg.GetAllowedPaths()

	for _, allowed := range allowedPaths {
		allowedUpper := strings.ToUpper(allowed)

		// Windows 盘符匹配
		if len(allowedUpper) >= 2 && allowedUpper[1] == ':' {
			allowedDrive := allowedUpper[:2]
			cleanDrive := ""
			if len(cleanPathUpper) >= 2 && cleanPathUpper[1] == ':' {
				cleanDrive = cleanPathUpper[:2]
			}

			if cleanDrive == allowedDrive {
				return true
			}
			continue
		}

		// Linux 路径匹配
		allowedClean := filepath.Clean(allowed)
		allowedCleanUpper := strings.ToUpper(allowedClean)

		if strings.HasPrefix(cleanPathUpper, allowedCleanUpper) {
			if len(cleanPath) == len(allowedClean) ||
				cleanPath[len(allowedClean)] == filepath.Separator {
				return true
			}
		}
	}

	logger.Warn("路径不在白名单中",
		zap.String("path", path),
		zap.String("cleaned", cleanPath),
		zap.Strings("allowed_paths", allowedPaths),
	)
	return false
}

func NewFileUpload(db *gorm.DB, redis *redis.Client, cfg *config.Config) _interface.FileUpload {
	return &fileUpload{
		BaseHandler: base.NewBaseHandler(db, redis),
		cfg:         cfg,
	}
}
