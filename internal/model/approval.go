package model

import "time"

// Approval 审批工单模型
type Approval struct {
	BaseModel
	Title       string     `gorm:"size:256;not null" json:"title"`               // 审批标题
	Description string     `gorm:"type:text" json:"description"`                 // 审批描述
	Type        string     `gorm:"size:64;not null;index" json:"type"`           // 审批类型: batch_delete / ssh_key_revoke / departure_cleanup / ssh_key_rotate
	Payload     string     `gorm:"type:text" json:"payload"`                     // 请求参数 JSON
	Status      string     `gorm:"size:32;default:pending;index" json:"status"`  // pending / approved / rejected / cancelled
	RequestedBy uint       `gorm:"index;not null" json:"requested_by"`           // 发起人 ID
	RequesterName string   `gorm:"size:128" json:"requester_name"`               // 发起人名称
	ReviewedBy  *uint      `json:"reviewed_by"`                                  // 审批人 ID
	ReviewerName string    `gorm:"size:128" json:"reviewer_name"`                // 审批人名称
	ReviewedAt  *time.Time `json:"reviewed_at"`                                  // 审批时间
	ReviewNote  string     `gorm:"size:512" json:"review_note"`                  // 审批备注
}
