package dto

// TerminalSessionResponse 终端会话响应
type TerminalSessionResponse struct {
	ID         uint   `json:"id"`
	AssetID    uint   `json:"asset_id"`
	Hostname   string `json:"hostname"`
	IP         string `json:"ip"`
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	Status     string `json:"status"`
	ClientIP   string `json:"client_ip"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
}

// ActiveSessionResponse 在线会话响应
type ActiveSessionResponse struct {
	SessionID uint   `json:"session_id"`
	AssetID   uint   `json:"asset_id"`
	Hostname  string `json:"hostname"`
	IP        string `json:"ip"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	ClientIP  string `json:"client_ip"`
	StartedAt string `json:"started_at"`
	Duration  string `json:"duration"` // 持续时长，如 "02:15:30"
}
