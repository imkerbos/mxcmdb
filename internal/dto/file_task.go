package dto

// FileDistributeRequest 文件分发请求
type FileDistributeRequest struct {
	Name       string `json:"name" binding:"required,max=128"`
	RemotePath string `json:"remote_path" binding:"required"`
	FileMode   string `json:"file_mode"`
	AssetIDs   []uint `json:"asset_ids" binding:"required"`
}

// FileTaskResponse 文件任务响应
type FileTaskResponse struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	FileName     string `json:"file_name"`
	FileSize     int64  `json:"file_size"`
	RemotePath   string `json:"remote_path"`
	FileMode     string `json:"file_mode"`
	Status       string `json:"status"`
	CreatedBy    uint   `json:"created_by"`
	TotalCount   int    `json:"total_count"`
	SuccessCount int    `json:"success_count"`
	FailCount    int    `json:"fail_count"`
	StartedAt    string `json:"started_at"`
	FinishedAt   string `json:"finished_at"`
	CreatedAt    string `json:"created_at"`
}

// FileTaskResultResponse 文件任务结果响应
type FileTaskResultResponse struct {
	ID         uint   `json:"id"`
	FileTaskID uint   `json:"file_task_id"`
	AssetID    uint   `json:"asset_id"`
	Hostname   string `json:"hostname"`
	IP         string `json:"ip"`
	Status     string `json:"status"`
	ErrorMsg   string `json:"error_msg"`
	Duration   int64  `json:"duration"`
}
