package dto

// SystemConfigListRequest 系统配置列表请求
type SystemConfigListRequest struct {
	Category string `form:"category"`
}

// SystemConfigUpdateRequest 系统配置更新请求
type SystemConfigUpdateRequest struct {
	Value string `json:"value" binding:"required"`
}

// SystemConfigBatchUpdateRequest 批量更新请求
type SystemConfigBatchUpdateRequest struct {
	Configs []SystemConfigItem `json:"configs" binding:"required,dive"`
}

// SystemConfigItem 单条配置
type SystemConfigItem struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// SystemConfigResponse 系统配置响应
type SystemConfigResponse struct {
	ID          uint   `json:"id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Type        string `json:"type"`
}
