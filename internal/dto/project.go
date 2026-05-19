package dto

// ProjectListRequest 项目列表请求
type ProjectListRequest struct {
	Keyword  string `form:"keyword"`
	Status   string `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required,max=64"`
	Code        string `json:"code" binding:"required,max=32"`
	Description string `json:"description" binding:"max=256"`
	Owner       string `json:"owner" binding:"max=64"`
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	Name        string `json:"name" binding:"max=64"`
	Description string `json:"description" binding:"max=256"`
	Owner       string `json:"owner" binding:"max=64"`
	Status      string `json:"status"`
}

// ProjectResponse 项目响应
type ProjectResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Owner       string `json:"owner"`
	Status      string `json:"status"`
	AssetCount  int64  `json:"asset_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ProjectSimple 项目简要信息（用于选择器）
type ProjectSimple struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// ProjectSummary 项目资产指纹聚合
type ProjectSummary struct {
	Project      ProjectResponse     `json:"project"`
	AssetStats   ProjectAssetStats   `json:"asset_stats"`
	StatusDist   []DistributionItem  `json:"status_distribution"`
	TypeDist     []DistributionItem  `json:"type_distribution"`
}

// ProjectAssetStats 项目资产聚合统计
type ProjectAssetStats struct {
	TotalAssets int64  `json:"total_assets"`
	TotalCPU    int64  `json:"total_cpu"`    // 总 CPU 核心数
	TotalMemory int64  `json:"total_memory"` // 总内存 (MB)
	TotalDisk   int64  `json:"total_disk"`   // 总磁盘 (GB)
	Probed      int64  `json:"probed"`       // 已探测数
	Running     int64  `json:"running"`      // 运行中数
	Stopped     int64  `json:"stopped"`      // 已停止数
}
