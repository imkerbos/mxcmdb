package model

import "time"

// SSHKey SSH 密钥模型
type SSHKey struct {
	BaseModel
	Name        string `gorm:"size:128;not null" json:"name"`
	PublicKey   string `gorm:"type:text;not null" json:"public_key"`
	PrivateKey  string `gorm:"type:text;not null" json:"-"`    // AES-GCM 加密存储
	Fingerprint string `gorm:"size:128" json:"fingerprint"`
	KeyType     string `gorm:"size:32" json:"key_type"`        // rsa / ed25519
	Comment     string `gorm:"size:256" json:"comment"`
	CreatedBy   uint   `json:"created_by"`
}

// SSHKeyBinding SSH 密钥绑定关系
type SSHKeyBinding struct {
	BaseModel
	SSHKeyID   uint       `gorm:"index:idx_binding_key_status,priority:1;not null" json:"ssh_key_id"`
	AssetID    uint       `gorm:"index;not null" json:"asset_id"`
	Username   string     `gorm:"size:64;not null;index:idx_binding_user_status,priority:1" json:"username"` // 目标主机用户名
	Status     string     `gorm:"size:32;default:pending;index:idx_binding_key_status,priority:2;index:idx_binding_user_status,priority:2" json:"status"` // deployed / pending / failed
	DeployedAt *time.Time `json:"deployed_at"`
	Asset      Asset      `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
}

// SSHKeyDeployLog SSH 密钥部署操作日志
type SSHKeyDeployLog struct {
	BaseModel
	SSHKeyID   uint   `gorm:"index;not null" json:"ssh_key_id"`
	SSHKeyName string `gorm:"size:128" json:"ssh_key_name"`
	AssetID    uint   `gorm:"index" json:"asset_id"`
	Hostname   string `gorm:"size:256" json:"hostname"`
	IP         string `gorm:"size:64" json:"ip"`
	Username   string `gorm:"size:64" json:"username"`
	Action     string `gorm:"size:32;not null" json:"action"` // deploy / revoke / rotate
	Status     string `gorm:"size:32;not null" json:"status"` // success / failed
	Error      string `gorm:"type:text" json:"error"`
	OperatorID uint   `gorm:"index" json:"operator_id"`
}
