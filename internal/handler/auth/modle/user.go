package model

import (
	"gorm.io/gorm"
	"time"
)

// User 用户表
type User struct {
	ID        uint       `gorm:"primaryKey;autoIncrement" json:"id"`                         // 用户ID
	Username  string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`      // 用户名
	Password  string     `gorm:"type:varchar(255);not null" json:"-"`                        // 密码（bcrypt加密），不返回给前端
	Email     string     `gorm:"type:varchar(100);uniqueIndex" json:"email"`                 // 邮箱
	Phone     string     `gorm:"type:varchar(20)" json:"phone"`                              // 手机号
	Avatar    string     `gorm:"type:varchar(255)" json:"avatar"`                            // 头像URL
	Role      string     `gorm:"type:varchar(20);default:user" json:"role"`                  // 角色：admin/user
	Status    int16      `gorm:"type:smallint;default:1;index" json:"status"`                // 状态：1正常/0禁用
	LastLogin *time.Time `gorm:"type:timestamp" json:"last_login"`                           // 最后登录时间
	CreatedAt time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"` // 创建时间
	UpdatedAt time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"` // 更新时间
}

// TableName 指定表名
func (User) TableName() string {
	return "user"
}

// 常量定义
const (
	RoleAdmin = "admin" // 管理员
	RoleUser  = "user"  // 普通用户
)

const (
	StatusActive   = 1 // 正常
	StatusInactive = 0 // 禁用
)

// BeforeCreate GORM钩子：创建前
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 可以在这里添加创建前的逻辑
	// 比如：设置默认值、验证等
	return nil
}

// BeforeUpdate GORM钩子：更新前
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// 更新时自动更新 updated_at
	u.UpdatedAt = time.Now()
	return nil
}
