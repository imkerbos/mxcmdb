package model

// Subnet 网段
type Subnet struct {
	BaseModel
	Name      string `gorm:"size:128;not null" json:"name"`
	CIDR      string `gorm:"size:64;not null;uniqueIndex" json:"cidr"`
	Gateway   string `gorm:"size:45" json:"gateway"`
	VLAN      int    `json:"vlan"`
	TotalIPs  int    `json:"total_ips"`
	UsedIPs   int    `json:"used_ips"`
	Comment   string `gorm:"size:256" json:"comment"`
}

// IPAddress IP 地址
type IPAddress struct {
	BaseModel
	SubnetID uint    `gorm:"index;not null" json:"subnet_id"`
	Subnet   Subnet  `gorm:"foreignKey:SubnetID" json:"subnet,omitempty"`
	Address  string  `gorm:"size:45;not null;uniqueIndex" json:"address"`
	Status   string  `gorm:"size:32;index;not null" json:"status"` // available / allocated / reserved
	AssetID  *uint   `gorm:"index" json:"asset_id"`
	Asset    *Asset  `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
	Hostname string  `gorm:"size:128" json:"hostname"`
	Comment  string  `gorm:"size:256" json:"comment"`
}
