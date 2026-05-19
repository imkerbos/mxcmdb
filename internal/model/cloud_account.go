package model

import "time"

// CloudAccount 云账号模型
type CloudAccount struct {
	BaseModel
	Name            string     `gorm:"size:128;not null" json:"name"`
	Provider        string     `gorm:"size:32;not null;default:aliyun" json:"provider"` // aliyun
	AccessKeyID     string     `gorm:"size:128;not null" json:"access_key_id"`
	AccessKeySecret string     `gorm:"size:512;not null" json:"-"` // AES-GCM 加密存储
	Region          string     `gorm:"size:64" json:"region"`
	Status          int        `gorm:"default:1" json:"status"` // 1=启用 0=禁用
	LastSyncAt      *time.Time `json:"last_sync_at"`
	LastSyncStatus  string     `gorm:"size:32" json:"last_sync_status"` // success / failed / syncing
	CreatedBy       uint       `json:"created_by"`
}
