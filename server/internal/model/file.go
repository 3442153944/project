package model

import (
	"gorm.io/gorm"
	"time"
)

// File 文件表
type File struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`                               // 文件ID
	UserID      uint       `gorm:"not null;index:idx_file_user" json:"user_id"`                      // 用户ID
	ParentID    *uint      `gorm:"index:idx_file_parent" json:"parent_id"`                           // 父目录ID
	FileName    string     `gorm:"type:varchar(255);not null" json:"file_name"`                      // 文件名
	FilePath    string     `gorm:"type:varchar(1000);not null;index:idx_file_path" json:"file_path"` // 文件完整路径
	FileType    string     `gorm:"type:varchar(20)" json:"file_type"`                                // 文件类型：doc/image/video/audio/other
	FileSize    int64      `gorm:"type:bigint" json:"file_size"`                                     // 文件大小（字节）
	FileHash    string     `gorm:"type:varchar(64);index" json:"file_hash"`                          // 文件哈希值（SHA256）
	MimeType    string     `gorm:"type:varchar(100)" json:"mime_type"`                               // MIME类型
	IsDirectory bool       `gorm:"type:boolean;default:false" json:"is_directory"`                   // 是否为目录
	IsDeleted   bool       `gorm:"type:boolean;default:false;index" json:"is_deleted"`               // 是否已删除
	Version     int        `gorm:"type:int;default:1" json:"version"`                                // 文件版本号
	ShareCode   string     `gorm:"type:varchar(32);index" json:"share_code,omitempty"`               // 分享码
	ShareExpire *time.Time `gorm:"type:timestamp" json:"share_expire,omitempty"`                     // 分享过期时间
	DeletedAt   *time.Time `gorm:"type:timestamp" json:"deleted_at,omitempty"`                       // 删除时间
	CreatedAt   time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`       // 创建时间
	UpdatedAt   time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`       // 更新时间

	// 关联
	User     *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`       // 所属用户
	Parent   *File  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`   // 父目录
	Children []File `gorm:"foreignKey:ParentID" json:"children,omitempty"` // 子文件/目录
}

// TableName 指定表名
func (File) TableName() string {
	return "file"
}

// 文件类型常量
const (
	FileTypeDoc   = "doc"   // 文档
	FileTypeImage = "image" // 图片
	FileTypeVideo = "video" // 视频
	FileTypeAudio = "audio" // 音频
	FileTypeOther = "other" // 其他
)

// BeforeCreate GORM钩子：创建前
func (f *File) BeforeCreate(tx *gorm.DB) error {
	// 可以在这里添加文件创建前的逻辑
	// 比如：自动识别文件类型、生成文件hash等
	return nil
}

// BeforeUpdate GORM钩子：更新前
func (f *File) BeforeUpdate(tx *gorm.DB) error {
	// 更新时自动更新 updated_at
	f.UpdatedAt = time.Now()
	return nil
}

// SoftDelete 软删除文件
func (f *File) SoftDelete(tx *gorm.DB) error {
	now := time.Now()
	f.IsDeleted = true
	f.DeletedAt = &now
	return tx.Model(f).Updates(map[string]interface{}{
		"is_deleted": true,
		"deleted_at": now,
	}).Error
}

// IsShared 判断文件是否已分享
func (f *File) IsShared() bool {
	return f.ShareCode != "" && (f.ShareExpire == nil || f.ShareExpire.After(time.Now()))
}

// FormatFileSize 格式化文件大小
func (f *File) FormatFileSize() string {
	return formatBytes(f.FileSize)
}
