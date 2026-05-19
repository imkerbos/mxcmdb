package model

// Notification 通知消息模型
type Notification struct {
	BaseModel
	UserID   uint   `gorm:"index;not null" json:"user_id"`   // 目标用户
	Title    string `gorm:"size:256;not null" json:"title"`   // 通知标题
	Content  string `gorm:"type:text" json:"content"`         // 通知内容
	Type     string `gorm:"size:32;default:info" json:"type"` // info / success / warning / error
	Source   string `gorm:"size:64" json:"source"`            // 来源模块: probe / task / ssh_key / terminal / system
	SourceID uint   `json:"source_id"`                        // 关联资源 ID
	Read     bool   `gorm:"default:false" json:"read"`        // 是否已读
}
