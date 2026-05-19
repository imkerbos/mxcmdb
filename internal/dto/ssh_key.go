package dto

// CreateSSHKeyRequest 创建 SSH Key 请求
type CreateSSHKeyRequest struct {
	Name       string `json:"name" binding:"required,max=128"`
	PublicKey  string `json:"public_key"`  // 导入时提供
	PrivateKey string `json:"private_key"` // 导入时提供
	KeyType    string `json:"key_type"`    // rsa / ed25519，生成时使用
	Comment    string `json:"comment"`
}

// DeploySSHKeyRequest 部署 SSH Key 请求
type DeploySSHKeyRequest struct {
	AssetIDs []uint `json:"asset_ids" binding:"required"`
	Username string `json:"username" binding:"required"`
}

// SSHKeyResponse SSH Key 响应
type SSHKeyResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	PublicKey   string `json:"public_key"`
	Fingerprint string `json:"fingerprint"`
	KeyType     string `json:"key_type"`
	Comment     string `json:"comment"`
	CreatedAt   string `json:"created_at"`
}

// SSHKeyBindingResponse SSH Key 绑定响应
type SSHKeyBindingResponse struct {
	ID         uint   `json:"id"`
	SSHKeyID   uint   `json:"ssh_key_id"`
	AssetID    uint   `json:"asset_id"`
	Hostname   string `json:"hostname"`
	IP         string `json:"ip"`
	Username   string `json:"username"`
	Status     string `json:"status"`
	DeployedAt string `json:"deployed_at"`
}

// RevokeSSHKeyRequest 撤销（回收）SSH Key 请求
type RevokeSSHKeyRequest struct {
	AssetIDs []uint `json:"asset_ids" binding:"required"` // 目标资产列表
	Username string `json:"username" binding:"required"`   // 要从哪个用户的 authorized_keys 移除
}

// RotateSSHKeyRequest SSH Key 轮换请求
type RotateSSHKeyRequest struct {
	NewKeyID   *uint  `json:"new_key_id"`  // 使用已有密钥替换，留空则自动生成
	NewName    string `json:"new_name"`    // 新密钥名称（自动生成时使用），留空则自动生成
	NewComment string `json:"new_comment"` // 新密钥备注（自动生成时使用）
}

// DepartureCleanupRequest 离职清理请求
type DepartureCleanupRequest struct {
	KeyIDs   []uint `json:"key_ids" binding:"required,min=1"` // 要清理的 SSH Key ID 列表
	AssetIDs []uint `json:"asset_ids"`                        // 指定资产列表，留空则清理所有已绑定资产
}

// RevokeResultItem 撤销结果项
type RevokeResultItem struct {
	AssetID  uint   `json:"asset_id"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Status   string `json:"status"` // success / failed
	Error    string `json:"error,omitempty"`
}

// RevokeResponse 撤销结果响应
type RevokeResponse struct {
	Total   int                `json:"total"`
	Success int                `json:"success"`
	Failed  int                `json:"failed"`
	Results []RevokeResultItem `json:"results"`
}

// RotateResponse SSH Key 轮换响应
type RotateResponse struct {
	OldKeyID uint            `json:"old_key_id"`
	NewKey   *SSHKeyResponse `json:"new_key"`
	Deploy   *RevokeResponse `json:"deploy"` // 新密钥部署结果
	Revoke   *RevokeResponse `json:"revoke"` // 旧密钥撤销结果
}

// DepartureCleanupResponse 离职清理响应
type DepartureCleanupResponse struct {
	TotalKeys     int              `json:"total_keys"`
	TotalAssets   int              `json:"total_assets"`
	RevokeResults []RevokeResponse `json:"revoke_results"`
}

// SSHKeyDownloadResponse SSH Key 下载响应（含解密后的私钥）
type SSHKeyDownloadResponse struct {
	Name       string `json:"name"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// SSHKeyDeployLogResponse 部署日志响应
type SSHKeyDeployLogResponse struct {
	ID         uint   `json:"id"`
	SSHKeyID   uint   `json:"ssh_key_id"`
	SSHKeyName string `json:"ssh_key_name"`
	AssetID    uint   `json:"asset_id"`
	Hostname   string `json:"hostname"`
	IP         string `json:"ip"`
	Username   string `json:"username"`
	Action     string `json:"action"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
	OperatorID uint   `json:"operator_id"`
	CreatedAt  string `json:"created_at"`
}
