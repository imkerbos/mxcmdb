package model

import "time"

// Asset 资产模型
type Asset struct {
	BaseModel
	Hostname       string     `gorm:"size:128" json:"hostname"`
	IP             string     `gorm:"size:45;uniqueIndex" json:"ip"`                   // 支持 IPv6
	Port           int        `gorm:"default:22" json:"port"`                           // SSH 端口
	OS             string     `gorm:"size:64" json:"os"`                                // 操作系统
	Type           string     `gorm:"size:32;index" json:"type"`                        // ecs/physical_server/vm/network_device 等
	Source         string     `gorm:"size:32;index" json:"source"`                      // cloud_sync/manual/probe
	CloudAccountID *uint      `gorm:"index" json:"cloud_account_id"`                    // 关联云账号
	InstanceID     string     `gorm:"size:128;index" json:"instance_id"`                // 云实例 ID
	Region         string     `gorm:"size:64" json:"region"`
	Zone           string     `gorm:"size:64" json:"zone"`
	Status         string     `gorm:"size:32;default:unknown;index" json:"status"`      // online/offline/unknown（IDC 探活）; running/stopped/terminated（云资源）
	Spec           string     `gorm:"size:128" json:"spec"`                             // CPU/内存规格
	Department     string     `gorm:"size:64;index" json:"department"`
	ProjectID      *uint      `gorm:"index" json:"project_id"`
	Owner          string     `gorm:"size:64" json:"owner"`
	Environment    string     `gorm:"size:32;index" json:"environment"`                 // dev/test/prod
	BusinessGroup  string     `gorm:"size:64" json:"business_group"`
	SshUser        string     `gorm:"size:64;default:root" json:"ssh_user"`
	SshKeyID       *uint      `json:"ssh_key_id"`
	SshPassword    string     `gorm:"size:512" json:"-"`                                // AES-GCM 加密存储
	ProbeLastAt    *time.Time `json:"probe_last_at"`
	Tags           []AssetTag `gorm:"foreignKey:AssetID" json:"tags"`
}

// AssetTag 资产标签模型
type AssetTag struct {
	BaseModel
	AssetID uint   `gorm:"uniqueIndex:idx_asset_tag;not null" json:"asset_id"`
	Key     string `gorm:"uniqueIndex:idx_asset_tag;size:64;not null" json:"key"`
	Value   string `gorm:"size:256;not null" json:"value"`
}
