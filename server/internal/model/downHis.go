package model

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

// DownloadHistory 下载历史记录表
type DownloadHistory struct {
	ID             uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint       `gorm:"not null;index:idx_download_user" json:"user_id"`
	DeviceID       uint       `gorm:"not null;index:idx_download_device" json:"device_id"`
	FileID         *uint      `gorm:"index:idx_download_file" json:"file_id"`
	FileName       string     `gorm:"type:varchar(255)" json:"file_name"`
	FileSize       int64      `gorm:"type:bigint" json:"file_size"`
	DownloadStatus string     `gorm:"type:varchar(20);default:pending" json:"download_status"`
	DownloadSpeed  int64      `gorm:"type:bigint" json:"download_speed"`
	IPAddress      string     `gorm:"type:varchar(50)" json:"ip_address"`
	StartedAt      *time.Time `gorm:"type:timestamp" json:"started_at"`
	CompletedAt    *time.Time `gorm:"type:timestamp" json:"completed_at"`
	CreatedAt      time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName 指定表名（GORM 特殊要求：值接收器）
func (DownloadHistory) TableName() string {
	return "download_history"
}

// 下载状态常量
const (
	DownloadStatusPending     = "pending"
	DownloadStatusDownloading = "downloading"
	DownloadStatusCompleted   = "completed"
	DownloadStatusFailed      = "failed"
	DownloadStatusCancelled   = "cancelled"
)

// ============ 以下所有方法统一使用指针接收器 ============

// BeforeCreate GORM钩子：创建前
func (dh *DownloadHistory) BeforeCreate(tx *gorm.DB) error {
	if dh.DownloadStatus == DownloadStatusDownloading && dh.StartedAt == nil {
		now := time.Now()
		dh.StartedAt = &now
	}
	return nil
}

// IsCompleted 判断是否下载完成
func (dh *DownloadHistory) IsCompleted() bool {
	return dh.DownloadStatus == DownloadStatusCompleted
}

// IsFailed 判断是否下载失败
func (dh *DownloadHistory) IsFailed() bool {
	return dh.DownloadStatus == DownloadStatusFailed
}

// IsCancelled 判断是否已取消
func (dh *DownloadHistory) IsCancelled() bool {
	return dh.DownloadStatus == DownloadStatusCancelled
}

// IsInProgress 判断是否正在下载
func (dh *DownloadHistory) IsInProgress() bool {
	return dh.DownloadStatus == DownloadStatusDownloading
}

// GetDuration 获取下载时长（秒）
func (dh *DownloadHistory) GetDuration() int64 {
	if dh.StartedAt == nil || dh.CompletedAt == nil {
		return 0
	}
	return int64(dh.CompletedAt.Sub(*dh.StartedAt).Seconds())
}

// GetAverageSpeed 获取平均下载速度（字节/秒）
func (dh *DownloadHistory) GetAverageSpeed() int64 {
	duration := dh.GetDuration()
	if duration == 0 {
		return 0
	}
	return dh.FileSize / duration
}

// FormatFileSize 格式化文件大小
func (dh *DownloadHistory) FormatFileSize() string {
	return formatBytes(dh.FileSize)
}

// FormatSpeed 格式化下载速度
func (dh *DownloadHistory) FormatSpeed() string {
	return formatBytes(dh.DownloadSpeed) + "/s"
}

// MarkAsCompleted 标记为已完成
func (dh *DownloadHistory) MarkAsCompleted(tx *gorm.DB) error {
	now := time.Now()
	dh.DownloadStatus = DownloadStatusCompleted
	dh.CompletedAt = &now
	return tx.Model(dh).Updates(map[string]interface{}{
		"download_status": DownloadStatusCompleted,
		"completed_at":    now,
	}).Error
}

// MarkAsFailed 标记为失败
func (dh *DownloadHistory) MarkAsFailed(tx *gorm.DB) error {
	now := time.Now()
	dh.DownloadStatus = DownloadStatusFailed
	dh.CompletedAt = &now
	return tx.Model(dh).Updates(map[string]interface{}{
		"download_status": DownloadStatusFailed,
		"completed_at":    now,
	}).Error
}

// formatBytes 字节格式化辅助函数
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
