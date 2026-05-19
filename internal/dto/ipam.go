package dto

// CreateSubnetRequest 创建网段请求
type CreateSubnetRequest struct {
	Name    string `json:"name" binding:"required,max=128"`
	CIDR    string `json:"cidr" binding:"required"`
	Gateway string `json:"gateway"`
	VLAN    int    `json:"vlan"`
	Comment string `json:"comment"`
}

// UpdateSubnetRequest 更新网段请求
type UpdateSubnetRequest struct {
	Name    string `json:"name"`
	Gateway string `json:"gateway"`
	VLAN    int    `json:"vlan"`
	Comment string `json:"comment"`
}

// SubnetResponse 网段响应
type SubnetResponse struct {
	ID       uint    `json:"id"`
	Name     string  `json:"name"`
	CIDR     string  `json:"cidr"`
	Gateway  string  `json:"gateway"`
	VLAN     int     `json:"vlan"`
	TotalIPs int     `json:"total_ips"`
	UsedIPs  int     `json:"used_ips"`
	UsagePercent float64 `json:"usage_percent"`
	Comment  string  `json:"comment"`
}

// AllocateIPRequest 分配 IP 请求
type AllocateIPRequest struct {
	AssetID  *uint  `json:"asset_id"`
	Hostname string `json:"hostname"`
	Comment  string `json:"comment"`
}

// IPAddressResponse IP 地址响应
type IPAddressResponse struct {
	ID       uint    `json:"id"`
	SubnetID uint    `json:"subnet_id"`
	Address  string  `json:"address"`
	Status   string  `json:"status"`
	AssetID  *uint   `json:"asset_id"`
	Hostname string  `json:"hostname"`
	Comment  string  `json:"comment"`
}
