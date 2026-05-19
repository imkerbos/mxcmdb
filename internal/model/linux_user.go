package model

// LinuxUser Linux 用户记录
type LinuxUser struct {
	BaseModel
	Username string `gorm:"size:64;not null;index" json:"username"`
	AssetID  uint   `gorm:"index;not null" json:"asset_id"`
	Asset    Asset  `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
	UID      int    `json:"uid"`
	GID      int    `json:"gid"`
	Home     string `gorm:"size:256" json:"home"`
	Shell    string `gorm:"size:128" json:"shell"`
	Sudo     bool   `json:"sudo"`
	Status   string `gorm:"size:32" json:"status"` // active / locked / deleted
}
