package dto

// DashboardStats 仪表盘统计
type DashboardStats struct {
	TotalAssets      int64              `json:"total_assets"`
	CloudAssets      int64              `json:"cloud_assets"`
	IDCAssets        int64              `json:"idc_assets"`
	OnlineAssets     int64              `json:"online_assets"`
	OfflineAssets    int64              `json:"offline_assets"`
	ProbedAssets     int64              `json:"probed_assets"`
	TotalProjects    int64              `json:"total_projects"`
	EnvDistribution  []DistributionItem `json:"env_distribution"`
	TypeDistribution []DistributionItem `json:"type_distribution"`
}

// DistributionItem 分布统计项
type DistributionItem struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

// RecentActivity 最近活动
type RecentActivity struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Module    string `json:"module"`
	Action    string `json:"action"`
	Path      string `json:"path"`
	ClientIP  string `json:"client_ip"`
	CreatedAt string `json:"created_at"`
}

// UpdateQuickActionsRequest 更新快捷操作请求
type UpdateQuickActionsRequest struct {
	Actions []string `json:"actions" binding:"required"`
}

// QuickActionsResponse 快捷操作响应
type QuickActionsResponse struct {
	Actions []string `json:"actions"`
}
