package model

import "time"

// User 平台用户模型
type User struct {
	BaseModel
	Username    string     `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Password    string     `gorm:"size:128;not null" json:"-"`
	Nickname    string     `gorm:"size:64" json:"nickname"`
	Email       string     `gorm:"size:128" json:"email"`
	Phone       string     `gorm:"size:32" json:"phone"`
	Avatar      string     `gorm:"size:256" json:"avatar"`
	Role        string     `gorm:"size:32;default:viewer" json:"role"` // admin / operator / viewer
	Status      int        `gorm:"default:1" json:"status"`            // 1=启用 0=禁用
	MFAEnabled  bool       `gorm:"default:false" json:"mfa_enabled"`
	MFASecret   string     `gorm:"size:64" json:"-"`
	LastLoginAt *time.Time `json:"last_login_at"`
	LastLoginIP string     `gorm:"size:45" json:"last_login_ip"`
	CreatedBy   uint       `json:"created_by"`
}
