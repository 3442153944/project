package model

import (
	"time"
)

type User struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Username  string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password  string     `gorm:"size:255;not null" json:"-"` //  不返回密码
	Email     string     `gorm:"uniqueIndex;size:100" json:"email"`
	Phone     string     `gorm:"size:20" json:"phone"`
	Avatar    string     `gorm:"size:255" json:"avatar"`
	Role      string     `gorm:"size:20;default:'user'" json:"role"`
	Status    int16      `gorm:"default:1" json:"status"` //  PostgreSQL用smallint
	LastLogin *time.Time `json:"last_login,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "user" //  匹配你的表名（注意没有s）
}
