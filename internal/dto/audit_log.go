package dto

// AuditLogListRequest 审计日志列表请求
type AuditLogListRequest struct {
	Module    string `form:"module"`
	Action    string `form:"action"`
	Username  string `form:"username"`
	StartDate string `form:"start_date"` // 格式: 2006-01-02
	EndDate   string `form:"end_date"`   // 格式: 2006-01-02
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}

// AuditLogResponse 审计日志响应
type AuditLogResponse struct {
	ID           uint   `json:"id"`
	CreatedAt    string `json:"created_at"`
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	Module       string `json:"module"`
	Action       string `json:"action"`
	Resource     string `json:"resource"`
	ResourceID   uint   `json:"resource_id"`
	ClientIP     string `json:"client_ip"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	RequestBody  string `json:"request_body"`
	ResponseCode int    `json:"response_code"`
	Duration     int64  `json:"duration"`
	Detail       string `json:"detail"`
}
