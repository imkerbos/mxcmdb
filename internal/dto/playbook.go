package dto

// CreatePlaybookRequest 创建剧本请求
type CreatePlaybookRequest struct {
	Name        string               `json:"name" binding:"required"`
	Description string               `json:"description"`
	Steps       []PlaybookStepConfig `json:"steps" binding:"required,min=1"`
}

// UpdatePlaybookRequest 更新剧本请求
type UpdatePlaybookRequest struct {
	Name        string               `json:"name" binding:"required"`
	Description string               `json:"description"`
	Enabled     *bool                `json:"enabled"`
	Steps       []PlaybookStepConfig `json:"steps" binding:"required,min=1"`
}

// PlaybookStepConfig 剧本步骤配置
type PlaybookStepConfig struct {
	Name    string `json:"name" binding:"required"`
	Script  string `json:"script" binding:"required"`
	Timeout int    `json:"timeout"` // 秒，默认 60
	OnError string `json:"on_error"` // stop / continue，默认 stop
}

// ExecutePlaybookRequest 执行剧本请求（支持多剧本合并执行）
type ExecutePlaybookRequest struct {
	PlaybookIDs []uint `json:"playbook_ids" binding:"required,min=1"`
	AssetIDs    []uint `json:"asset_ids" binding:"required,min=1"`
}

// PlaybookJobResponse 异步执行任务引用
type PlaybookJobResponse struct {
	JobID       string `json:"job_id"`
	ExecutionID uint   `json:"execution_id"`
	Total       int    `json:"total"`
}

// PlaybookResponse 剧本详情响应
type PlaybookResponse struct {
	ID          uint                 `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Category    string               `json:"category"`
	IsBuiltin   bool                 `json:"is_builtin"`
	Enabled     bool                 `json:"enabled"`
	Steps       []PlaybookStepConfig `json:"steps"`
	CreatedBy   uint                 `json:"created_by"`
	CreatedAt   string               `json:"created_at"`
	UpdatedAt   string               `json:"updated_at"`
}

// PlaybookExecuteResultItem 单资产执行结果
type PlaybookExecuteResultItem struct {
	AssetID     uint              `json:"asset_id"`
	Hostname    string            `json:"hostname"`
	IP          string            `json:"ip"`
	Status      string            `json:"status"` // success / partial / failed
	StepResults []StepExecResult  `json:"step_results"`
}

// StepExecResult 单步骤执行结果
type StepExecResult struct {
	Name     string `json:"name"`
	Status   string `json:"status"` // success / failed / skipped
	Output   string `json:"output,omitempty"`
	Error    string `json:"error,omitempty"`
	Latency  int64  `json:"latency"` // 毫秒
}

// PlaybookExecuteResponse 执行完成响应
type PlaybookExecuteResponse struct {
	Total   int                         `json:"total"`
	Success int                         `json:"success"`
	Failed  int                         `json:"failed"`
	Results []PlaybookExecuteResultItem `json:"results"`
}

// PlaybookExecutionListItem 执行历史列表项
type PlaybookExecutionListItem struct {
	ID           uint   `json:"id"`
	PlaybookIDs  string `json:"playbook_ids"`
	PlaybookName string `json:"playbook_name"`
	Status       string `json:"status"`
	TotalAssets  int    `json:"total_assets"`
	Success      int    `json:"success"`
	Failed       int    `json:"failed"`
	StartedAt    string `json:"started_at"`
	FinishedAt   string `json:"finished_at"`
	CreatedBy    uint   `json:"created_by"`
}
