package model

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

// UploadHistory 文件上传记录表
type UploadHistory struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint       `gorm:"not null;index:idx_upload_user" json:"user_id"`
	FileName     string     `gorm:"type:varchar(255);not null" json:"file_name"`
	OriginalName string     `gorm:"type:varchar(255)" json:"original_name"`
	FileSize     int64      `gorm:"type:bigint" json:"file_size"`
	FileType     string     `gorm:"type:varchar(100)" json:"file_type"`
	StoragePath  string     `gorm:"type:varchar(500)" json:"storage_path"`
	UploadStatus string     `gorm:"type:varchar(20);default:pending;index:idx_upload_status" json:"upload_status"`
	UploadSpeed  int64      `gorm:"type:bigint" json:"upload_speed"`
	Progress     int        `gorm:"type:integer;default:0" json:"progress"`
	IPAddress    string     `gorm:"type:varchar(50)" json:"ip_address"`
	UserAgent    string     `gorm:"type:varchar(500)" json:"user_agent"`
	ErrorMessage string     `gorm:"type:text" json:"error_message"`
	StartedAt    *time.Time `gorm:"type:timestamp" json:"started_at"`
	CompletedAt  *time.Time `gorm:"type:timestamp" json:"completed_at"`
	CreatedAt    time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;index:idx_upload_created" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 指定表名
func (UploadHistory) TableName() string {
	return "upload_history"
}

// 上传状态常量
const (
	UploadStatusPending   = "pending"
	UploadStatusUploading = "uploading"
	UploadStatusCompleted = "completed"
	UploadStatusFailed    = "failed"
	UploadStatusCancelled = "cancelled"
)

// ============ 以下所有方法统一使用指针接收器 ============

// BeforeCreate GORM钩子：创建前
func (uh *UploadHistory) BeforeCreate(tx *gorm.DB) error {
	if uh.UploadStatus == UploadStatusUploading && uh.StartedAt == nil {
		now := time.Now()
		uh.StartedAt = &now
	}
	return nil
}

// IsCompleted 判断是否上传完成
func (uh *UploadHistory) IsCompleted() bool {
	return uh.UploadStatus == UploadStatusCompleted
}

// IsFailed 判断是否上传失败
func (uh *UploadHistory) IsFailed() bool {
	return uh.UploadStatus == UploadStatusFailed
}

// IsCancelled 判断是否已取消
func (uh *UploadHistory) IsCancelled() bool {
	return uh.UploadStatus == UploadStatusCancelled
}

// IsInProgress 判断是否正在上传
func (uh *UploadHistory) IsInProgress() bool {
	return uh.UploadStatus == UploadStatusUploading
}

// GetDuration 获取上传时长（秒）
func (uh *UploadHistory) GetDuration() int64 {
	if uh.StartedAt == nil || uh.CompletedAt == nil {
		return 0
	}
	return int64(uh.CompletedAt.Sub(*uh.StartedAt).Seconds())
}

// GetAverageSpeed 获取平均上传速度（字节/秒）
func (uh *UploadHistory) GetAverageSpeed() int64 {
	duration := uh.GetDuration()
	if duration == 0 {
		return 0
	}
	return uh.FileSize / duration
}

// FormatFileSize 格式化文件大小
func (uh *UploadHistory) FormatFileSize() string {
	return formatBytes(uh.FileSize)
}

// FormatSpeed 格式化上传速度
func (uh *UploadHistory) FormatSpeed() string {
	return formatBytes(uh.UploadSpeed) + "/s"
}

// MarkAsUploading 标记为上传中
func (uh *UploadHistory) MarkAsUploading(tx *gorm.DB) error {
	now := time.Now()
	uh.UploadStatus = UploadStatusUploading
	uh.StartedAt = &now
	return tx.Model(uh).Updates(map[string]interface{}{
		"upload_status": UploadStatusUploading,
		"started_at":    now,
	}).Error
}

// UpdateProgress 更新上传进度
func (uh *UploadHistory) UpdateProgress(tx *gorm.DB, progress int, speed int64) error {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	uh.Progress = progress
	uh.UploadSpeed = speed
	return tx.Model(uh).Updates(map[string]interface{}{
		"progress":     progress,
		"upload_speed": speed,
	}).Error
}

// MarkAsCompleted 标记为已完成
func (uh *UploadHistory) MarkAsCompleted(tx *gorm.DB) error {
	now := time.Now()
	uh.UploadStatus = UploadStatusCompleted
	uh.CompletedAt = &now
	uh.Progress = 100
	return tx.Model(uh).Updates(map[string]interface{}{
		"upload_status": UploadStatusCompleted,
		"completed_at":  now,
		"progress":      100,
	}).Error
}

// MarkAsFailed 标记为失败
func (uh *UploadHistory) MarkAsFailed(tx *gorm.DB, errMsg string) error {
	now := time.Now()
	uh.UploadStatus = UploadStatusFailed
	uh.CompletedAt = &now
	uh.ErrorMessage = errMsg
	return tx.Model(uh).Updates(map[string]interface{}{
		"upload_status": UploadStatusFailed,
		"completed_at":  now,
		"error_message": errMsg,
	}).Error
}

// MarkAsCancelled 标记为已取消
func (uh *UploadHistory) MarkAsCancelled(tx *gorm.DB) error {
	now := time.Now()
	uh.UploadStatus = UploadStatusCancelled
	uh.CompletedAt = &now
	return tx.Model(uh).Updates(map[string]interface{}{
		"upload_status": UploadStatusCancelled,
		"completed_at":  now,
	}).Error
}

// GetProgressPercentage 获取进度百分比字符串
func (uh *UploadHistory) GetProgressPercentage() string {
	return fmt.Sprintf("%d%%", uh.Progress)
}
