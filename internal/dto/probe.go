package dto

// ProbeRequest 探针执行请求
type ProbeRequest struct {
	AssetIDs []uint `json:"asset_ids" binding:"required"`
}

// ProbeResultResponse 探针结果响应
type ProbeResultResponse struct {
	ID              uint   `json:"id"`
	AssetID         uint   `json:"asset_id"`
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
	CollectedAt     string `json:"collected_at"`
	Status          string `json:"status"`
	ErrorMessage    string `json:"error_message"`
}
