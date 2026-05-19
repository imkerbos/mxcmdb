package dto

// CreateLinuxUserRequest 创建 Linux 用户请求
type CreateLinuxUserRequest struct {
	Username string `json:"username" binding:"required,max=64"`
	AssetIDs []uint `json:"asset_ids" binding:"required"`
	Shell    string `json:"shell"`
	Sudo     bool   `json:"sudo"`
}

// DeleteLinuxUserRequest 删除 Linux 用户请求
type DeleteLinuxUserRequest struct {
	AssetIDs []uint `json:"asset_ids" binding:"required"`
}

// LinuxUserResponse Linux 用户响应
type LinuxUserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	AssetID  uint   `json:"asset_id"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	UID      int    `json:"uid"`
	GID      int    `json:"gid"`
	Home     string `json:"home"`
	Shell    string `json:"shell"`
	Sudo     bool   `json:"sudo"`
	Status   string `json:"status"`
}
