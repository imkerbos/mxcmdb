package dto

// AssetListRequest 资产列表请求
type AssetListRequest struct {
	Keyword     string `form:"keyword"`
	Type        string `form:"type"`
	Source      string `form:"source"`
	Status      string `form:"status"`
	Environment string `form:"environment"`
	Department  string `form:"department"`
	Project     string `form:"project"`
	Owner       string `form:"owner"`
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
}

// CreateAssetRequest 创建资产请求
type CreateAssetRequest struct {
	Hostname      string          `json:"hostname" binding:"max=128"`
	IP            string          `json:"ip" binding:"required"`
	Port          int             `json:"port"`
	OS            string          `json:"os"`
	Type          string          `json:"type" binding:"required"`
	Source        string          `json:"source"`
	Status        string          `json:"status"`
	Spec          string          `json:"spec"`
	Department    string          `json:"department"`
	ProjectID     *uint           `json:"project_id"`
	Owner         string          `json:"owner"`
	Environment   string          `json:"environment"`
	BusinessGroup string          `json:"business_group"`
	SshUser       string          `json:"ssh_user"`
	SshPassword   string          `json:"ssh_password"`
	SshKeyID      *uint           `json:"ssh_key_id"`
	Tags          []AssetTagItem  `json:"tags"`
}

// UpdateAssetRequest 更新资产请求
type UpdateAssetRequest struct {
	Hostname      string          `json:"hostname"`
	Port          int             `json:"port"`
	OS            string          `json:"os"`
	Type          string          `json:"type"`
	Status        string          `json:"status"`
	Spec          string          `json:"spec"`
	Department    string          `json:"department"`
	ProjectID     *uint           `json:"project_id"`
	Owner         string          `json:"owner"`
	Environment   string          `json:"environment"`
	BusinessGroup string          `json:"business_group"`
	SshUser       string          `json:"ssh_user"`
	SshPassword   string          `json:"ssh_password"`
	SshKeyID      *uint           `json:"ssh_key_id"`
	ClearSshKey   bool            `json:"clear_ssh_key"`
	Tags          []AssetTagItem  `json:"tags"`
}

// AssetTagItem 标签项
type AssetTagItem struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// AssetResponse 资产响应
type AssetResponse struct {
	ID             uint           `json:"id"`
	Hostname       string         `json:"hostname"`
	IP             string         `json:"ip"`
	Port           int            `json:"port"`
	OS             string         `json:"os"`
	Type           string         `json:"type"`
	Source         string         `json:"source"`
	CloudAccountID *uint          `json:"cloud_account_id"`
	InstanceID     string         `json:"instance_id"`
	Region         string         `json:"region"`
	Zone           string         `json:"zone"`
	Status         string         `json:"status"`
	Spec           string         `json:"spec"`
	Department     string         `json:"department"`
	ProjectID      *uint          `json:"project_id"`
	ProjectName    string         `json:"project_name"`
	Owner          string         `json:"owner"`
	Environment    string         `json:"environment"`
	BusinessGroup  string         `json:"business_group"`
	SshUser        string         `json:"ssh_user"`
	SshKeyID       *uint          `json:"ssh_key_id"`
	SshKeyName     string         `json:"ssh_key_name"`
	HasPassword    bool           `json:"has_password"`
	ProbeLastAt    string         `json:"probe_last_at"`
	Tags           []AssetTagItem `json:"tags"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
}

// BatchImportAssetRequest 批量导入资产请求
type BatchImportAssetRequest struct {
	Assets []CreateAssetRequest `json:"assets" binding:"required,min=1,max=500"`
}

// BatchImportAssetResponse 批量导入响应
type BatchImportAssetResponse struct {
	Total   int                     `json:"total"`
	Success int                     `json:"success"`
	Failed  int                     `json:"failed"`
	Errors  []BatchImportErrorItem  `json:"errors,omitempty"`
}

// BatchImportErrorItem 导入失败项
type BatchImportErrorItem struct {
	Index   int    `json:"index"`
	IP      string `json:"ip"`
	Message string `json:"message"`
}

// BatchTestConnectionRequest 批量测试连接请求
type BatchTestConnectionRequest struct {
	AssetIDs []uint `json:"asset_ids" binding:"required,min=1,max=100"`
}

// BatchTestConnectionResponse 批量测试连接响应
type BatchTestConnectionResponse struct {
	Results []TestConnectionResult `json:"results"`
}

// TestConnectionResult 单个连接测试结果
type TestConnectionResult struct {
	AssetID    uint   `json:"asset_id"`
	Hostname   string `json:"hostname"`
	IP         string `json:"ip"`
	Status     string `json:"status"`
	Latency    int64  `json:"latency"`
	AuthMethod string `json:"auth_method"` // password / key / password+key
	Error      string `json:"error,omitempty"`
}

// BatchDeleteAssetRequest 批量删除资产请求
type BatchDeleteAssetRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1,max=500"`
}

// BatchDeleteAssetResponse 批量删除响应
type BatchDeleteAssetResponse struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

// AssetDetailResponse 资产详情（聚合所有关联数据）
type AssetDetailResponse struct {
	// 基本信息
	AssetResponse

	// 最新探针数据
	Probe *AssetDetailProbe `json:"probe"`

	// 关联信息
	SSHKeyBindings   []AssetDetailSSHKeyBinding   `json:"ssh_key_bindings"`
	LinuxUsers       []AssetDetailLinuxUser        `json:"linux_users"`
	TerminalSessions []AssetDetailTerminalSession  `json:"terminal_sessions"`
	ProbeHistory     []AssetDetailProbeHistory     `json:"probe_history"`
}

// AssetDetailProbe 探针详情数据
type AssetDetailProbe struct {
	Hostname        string `json:"hostname"`
	CPU             string `json:"cpu"`
	Memory          string `json:"memory"`
	Disk            string `json:"disk"`
	Network         string `json:"network"`
	OS              string `json:"os"`
	Kernel          string `json:"kernel"`
	DockerVersion   string `json:"docker_version"`
	RunningServices string `json:"running_services"`
	SSHUsers        string `json:"ssh_users"`
	Uptime          string `json:"uptime"`
	DNS             string `json:"dns"`
	Gateway         string `json:"gateway"`
	Manufacturer    string `json:"manufacturer"`
	ProductModel    string `json:"product_model"`
	SerialNumber    string `json:"serial_number"`
	PublicIP        string `json:"public_ip"`
	Processes       string `json:"processes"`
	Listeners       string `json:"listeners"`
	CollectedAt     string `json:"collected_at"`
	Status          string `json:"status"`
}

// AssetDetailSSHKeyBinding SSH Key 绑定信息
type AssetDetailSSHKeyBinding struct {
	ID         uint   `json:"id"`
	SSHKeyID   uint   `json:"ssh_key_id"`
	KeyName    string `json:"key_name"`
	Username   string `json:"username"`
	Status     string `json:"status"`
	DeployedAt string `json:"deployed_at"`
}

// AssetDetailLinuxUser Linux 用户信息
type AssetDetailLinuxUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	UID      int    `json:"uid"`
	GID      int    `json:"gid"`
	Home     string `json:"home"`
	Shell    string `json:"shell"`
	Sudo     bool   `json:"sudo"`
	Status   string `json:"status"`
}

// AssetDetailTerminalSession 终端会话记录
type AssetDetailTerminalSession struct {
	ID         uint   `json:"id"`
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	Status     string `json:"status"`
	ClientIP   string `json:"client_ip"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
}

// AssetDetailProbeHistory 探针历史记录
type AssetDetailProbeHistory struct {
	ID          uint   `json:"id"`
	Status      string `json:"status"`
	OS          string `json:"os"`
	Kernel      string `json:"kernel"`
	CollectedAt string `json:"collected_at"`
}

// AssetStatsResponse 资产统计
type AssetStatsResponse struct {
	TotalCloud    int64 `json:"total_cloud"`
	TotalIDC      int64 `json:"total_idc"`
	TotalOnline   int64 `json:"total_online"`
	TotalOffline  int64 `json:"total_offline"`
	TotalProbed   int64 `json:"total_probed"`
}
