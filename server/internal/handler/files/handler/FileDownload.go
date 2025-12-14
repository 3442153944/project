// handler/files/handler/get_file.go
package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

type getFile struct {
	*base.BaseHandler
	cfg *config.Config
}

// HandlerGET 处理文件下载请求（GET 方法）
func (g *getFile) HandlerGET(c *gin.Context) {
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

	// 3. 从 Query 参数获取
	path := c.Query("path")
	name := c.Query("name")
	deviceIDStr := c.Query("device_id")

	// 4. 验证参数
	if path == "" || name == "" {
		response.BadRequest(c, "缺少必要参数 path 或 name")
		return
	}

	var deviceID uint
	if deviceIDStr != "" {
		if id, err := strconv.ParseUint(deviceIDStr, 10, 32); err == nil {
			deviceID = uint(id)
		}
	}

	// 5. 智能构建完整路径
	var fullPath string

	// 检查 path 是否已经是完整路径（包含文件名）
	if filepath.Base(path) == name {
		// path 已经包含文件名，直接使用
		fullPath = path
		logger.Info("使用完整路径",
			zap.String("path", path),
			zap.String("name", name),
		)
	} else {
		// path 只是目录，需要拼接文件名
		fullPath = filepath.Join(path, name)
		logger.Info("拼接路径",
			zap.String("dir", path),
			zap.String("name", name),
			zap.String("full_path", fullPath),
		)
	}

	// 6. 验证路径安全性
	if !g.isPathAllowed(fullPath) {
		logger.Warn("路径访问被拒绝",
			zap.Uint("user_id", userID),
			zap.String("path", fullPath),
		)
		response.Forbidden(c, "无权访问该路径")
		return
	}

	// 7. 获取文件信息
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warn("文件不存在",
				zap.String("full_path", fullPath),
				zap.String("path_param", path),
				zap.String("name_param", name),
			)
			response.NotFound(c, "文件不存在")
		} else {
			logger.Error("获取文件信息失败",
				zap.Error(err),
				zap.String("path", fullPath),
			)
			response.InternalError(c, "获取文件信息失败: "+err.Error())
		}
		return
	}

	// 8. 检查是否为目录
	if fileInfo.IsDir() {
		response.BadRequest(c, "不能下载目录")
		return
	}

	fileSize := fileInfo.Size()

	// 9. 创建下载历史记录
	history := &model.DownloadHistory{
		UserID:         userID,
		DeviceID:       deviceID,
		FileName:       name,
		FileSize:       fileSize,
		DownloadStatus: model.DownloadStatusPending,
		IPAddress:      c.ClientIP(),
	}

	if err := g.DB.Create(history).Error; err != nil {
		logger.Error("创建下载历史失败",
			zap.Error(err),
			zap.Uint("user_id", userID),
		)
		// 不影响主流程
	}

	// 10. 打开文件
	file, err := os.Open(fullPath)
	if err != nil {
		logger.Error("打开文件失败",
			zap.Error(err),
			zap.String("path", fullPath),
		)
		response.InternalError(c, "打开文件失败")
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// 11. 处理 Range 请求（断点续传）
	rangeHeader := c.GetHeader("Range")

	if rangeHeader == "" {
		// 没有 Range，返回完整文件
		g.serveFullFile(c, file, fileInfo, name, userID, history.ID)
	} else {
		// 有 Range，返回部分文件
		g.serveRangeFile(c, file, fileSize, rangeHeader, name, userID, history.ID)
	}
}

// serveFullFile 返回完整文件
func (g *getFile) serveFullFile(c *gin.Context, file *os.File, fileInfo os.FileInfo, fileName string, userID uint, historyID uint) {
	c.Header("Content-Type", getMimeType(fileName))
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Accept-Ranges", "bytes")
	c.Header("X-History-ID", strconv.FormatUint(uint64(historyID), 10))

	if historyID > 0 {
		_ = g.DB.Model(&model.DownloadHistory{}).
			Where("id = ?", historyID).
			Update("download_status", model.DownloadStatusDownloading).Error
	}

	_ = websocket.SendToUser(userID, "file_download", map[string]interface{}{
		"event":      "start",
		"file_name":  fileName,
		"file_size":  fileInfo.Size(),
		"history_id": historyID,
	})

	c.Status(http.StatusOK)

	written, err := io.Copy(c.Writer, file)
	if err != nil {
		logger.Error("文件传输失败", zap.Error(err))
		_ = g.DB.Model(&model.DownloadHistory{}).
			Where("id = ?", historyID).
			Update("download_status", model.DownloadStatusFailed).Error
		return
	}

	if historyID > 0 {
		_ = g.DB.Model(&model.DownloadHistory{}).
			Where("id = ?", historyID).
			Updates(map[string]interface{}{
				"download_status": model.DownloadStatusCompleted,
				"completed_at":    gorm.Expr("NOW()"),
			}).Error
	}

	_ = websocket.SendToUser(userID, "file_download", map[string]interface{}{
		"event":      "completed",
		"file_name":  fileName,
		"file_size":  written,
		"history_id": historyID,
	})

	logger.Info("文件下载完成",
		zap.Uint("user_id", userID),
		zap.String("file_name", fileName),
		zap.Int64("file_size", written),
	)
}

// serveRangeFile 返回部分文件（断点续传）
func (g *getFile) serveRangeFile(c *gin.Context, file *os.File, fileSize int64, rangeHeader string, fileName string, userID uint, historyID uint) {
	ranges := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(ranges, "-")

	if len(parts) != 2 {
		response.BadRequest(c, "无效的 Range 请求")
		return
	}

	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || start < 0 || start >= fileSize {
		c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		response.Error(c, http.StatusRequestedRangeNotSatisfiable, "Range 超出范围")
		return
	}

	var end int64
	if parts[1] == "" {
		end = fileSize - 1
	} else {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil || end >= fileSize {
			end = fileSize - 1
		}
	}

	if start > end {
		c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		response.Error(c, http.StatusRequestedRangeNotSatisfiable, "Range 起始位置大于结束位置")
		return
	}

	contentLength := end - start + 1

	if _, err := file.Seek(start, io.SeekStart); err != nil {
		logger.Error("文件定位失败", zap.Error(err))
		response.InternalError(c, "文件定位失败")
		return
	}

	c.Header("Content-Type", getMimeType(fileName))
	c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Accept-Ranges", "bytes")
	c.Header("X-History-ID", strconv.FormatUint(uint64(historyID), 10))

	if historyID > 0 {
		_ = g.DB.Model(&model.DownloadHistory{}).
			Where("id = ?", historyID).
			Update("download_status", model.DownloadStatusDownloading).Error
	}

	c.Status(http.StatusPartialContent)

	written, err := io.CopyN(c.Writer, file, contentLength)
	if err != nil && err != io.EOF {
		logger.Error("文件传输失败", zap.Error(err))
		return
	}

	if end == fileSize-1 {
		if historyID > 0 {
			_ = g.DB.Model(&model.DownloadHistory{}).
				Where("id = ?", historyID).
				Updates(map[string]interface{}{
					"download_status": model.DownloadStatusCompleted,
					"completed_at":    gorm.Expr("NOW()"),
				}).Error
		}

		_ = websocket.SendToUser(userID, "file_download", map[string]interface{}{
			"event":      "completed",
			"file_name":  fileName,
			"file_size":  fileSize,
			"history_id": historyID,
		})
	}

	logger.Info("Range 下载完成",
		zap.Uint("user_id", userID),
		zap.String("file_name", fileName),
		zap.Int64("start", start),
		zap.Int64("end", end),
		zap.Int64("written", written),
	)
}

// isPathAllowed 路径安全检查
func (g *getFile) isPathAllowed(path string) bool {
	cleanPath := filepath.Clean(path)
	cleanPathUpper := strings.ToUpper(cleanPath)

	if !filepath.IsAbs(cleanPath) {
		logger.Warn("拒绝相对路径", zap.String("path", path))
		return false
	}

	if strings.Contains(path, "..") {
		logger.Warn("拒绝包含..的路径", zap.String("path", path))
		return false
	}

	allowedPaths := g.cfg.GetAllowedPaths()

	for _, allowed := range allowedPaths {
		allowedUpper := strings.ToUpper(allowed)

		if len(allowedUpper) == 2 && allowedUpper[1] == ':' {
			if strings.HasPrefix(cleanPathUpper, allowedUpper) {
				return true
			}
			continue
		}

		allowedClean := filepath.Clean(allowed)
		allowedCleanUpper := strings.ToUpper(allowedClean)

		if strings.HasPrefix(cleanPathUpper, allowedCleanUpper) {
			if len(cleanPath) == len(allowedClean) ||
				cleanPath[len(allowedClean)] == filepath.Separator {
				return true
			}
		}
	}

	logger.Warn("路径不在白名单中", zap.String("path", path))
	return false
}

// getMimeType 获取 MIME 类型
func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	mimeTypes := map[string]string{
		".txt":  "text/plain",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		".rar":  "application/x-rar-compressed",
		".7z":   "application/x-7z-compressed",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".mp4":  "video/mp4",
		".mp3":  "audio/mpeg",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}

func NewGetFile(db *gorm.DB, redis *redis.Client, cfg *config.Config) _interface.GetFile {
	return &getFile{
		BaseHandler: base.NewBaseHandler(db, redis),
		cfg:         cfg,
	}
}
