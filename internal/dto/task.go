package dto

// TaskExecuteRequest 批量执行请求
type TaskExecuteRequest struct {
	Name     string `json:"name" binding:"required,max=128"`
	Command  string `json:"command" binding:"required"`
	AssetIDs []uint `json:"asset_ids" binding:"required"`
}

// TaskListRequest 任务列表请求
type TaskListRequest struct {
	Status string `form:"status"`
	Page   int    `form:"page"`
	Size   int    `form:"page_size"`
}

// TaskResponse 任务响应
type TaskResponse struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Command       string `json:"command"`
	Status        string `json:"status"`
	CreatedBy     uint   `json:"created_by"`
	CreatedByName string `json:"created_by_name"`
	TotalCount    int    `json:"total_count"`
	SuccessCount  int    `json:"success_count"`
	FailCount     int    `json:"fail_count"`
	StartedAt     string `json:"started_at"`
	FinishedAt    string `json:"finished_at"`
	CreatedAt     string `json:"created_at"`
}

// TaskResultResponse 任务结果响应
type TaskResultResponse struct {
	ID       uint   `json:"id"`
	TaskID   uint   `json:"task_id"`
	AssetID  uint   `json:"asset_id"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Status   string `json:"status"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Duration int64  `json:"duration"`
}
