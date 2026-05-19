package dto

// CreateCloudAccountRequest 创建云账号请求
type CreateCloudAccountRequest struct {
	Name            string `json:"name" binding:"required,max=128"`
	Provider        string `json:"provider" binding:"required,oneof=aliyun"`
	AccessKeyID     string `json:"access_key_id" binding:"required"`
	AccessKeySecret string `json:"access_key_secret" binding:"required"`
	Region          string `json:"region"`
}

// UpdateCloudAccountRequest 更新云账号请求
type UpdateCloudAccountRequest struct {
	Name            string `json:"name" binding:"max=128"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"` // 为空表示不更新
	Region          string `json:"region"`
}

// CloudAccountResponse 云账号响应
type CloudAccountResponse struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Provider        string `json:"provider"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"` // 脱敏显示
	Region          string `json:"region"`
	Status          int    `json:"status"`
	LastSyncAt      string `json:"last_sync_at"`
	LastSyncStatus  string `json:"last_sync_status"`
	CreatedAt       string `json:"created_at"`
}

// CloudAccountListRequest 云账号列表请求
type CloudAccountListRequest struct {
	Keyword  string `form:"keyword"`
	Provider string `form:"provider"`
	Status   *int   `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}
