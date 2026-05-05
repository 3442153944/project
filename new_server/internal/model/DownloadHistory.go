package model

import "time"

// DownloadHistory 下载历史记录表
type DownloadHistory struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint       `gorm:"not null;index" json:"user_id"`
	DeviceID       uint       `gorm:"not null;index" json:"device_id"`
	FileID         *uint64    `json:"file_id"`
	FileName       *string    `gorm:"size:255" json:"file_name"`
	FileSize       *int64     `json:"file_size"`
	DownloadStatus string     `gorm:"size:20;not null;default:pending" json:"download_status"`
	DownloadSpeed  *int64     `json:"download_speed"`
	IPAddress      *string    `gorm:"size:50" json:"ip_address"`
	StartedAt      *time.Time `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	CreatedAt      time.Time  `gorm:"autoCreateTime;index" json:"created_at"`
}

func (DownloadHistory) TableName() string { return "download_history" }
