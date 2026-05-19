package model

import "time"

// AuditLog 审计日志模型
type AuditLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time `gorm:"index" json:"created_at"`
	UserID       uint      `gorm:"index" json:"user_id"`
	Username     string    `gorm:"size:64" json:"username"`
	Module       string    `gorm:"size:32;index" json:"module"`   // auth / asset / cloud / ssh / task / file / terminal / user / ipam / settings
	Action       string    `gorm:"size:32;index" json:"action"`   // login / create / update / delete / execute / deploy / sync
	Resource     string    `gorm:"size:64" json:"resource"`       // 资源类型
	ResourceID   uint      `json:"resource_id"`                   // 资源 ID
	ClientIP     string    `gorm:"size:45" json:"client_ip"`      // 支持 IPv6
	Method       string    `gorm:"size:10" json:"method"`         // HTTP Method
	Path         string    `gorm:"size:256" json:"path"`          // 请求路径
	RequestBody  string    `gorm:"type:text" json:"request_body"` // 请求体（脱敏）
	ResponseCode int       `json:"response_code"`                 // 业务响应码
	Duration     int64     `json:"duration"`                      // 耗时（毫秒）
	Detail       string    `gorm:"type:text" json:"detail"`       // 附加信息
}
