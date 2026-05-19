package model

// UserQuickAction 用户快捷操作配置
type UserQuickAction struct {
	BaseModel
	UserID  uint   `gorm:"uniqueIndex;not null" json:"user_id"`
	Actions string `gorm:"type:text;not null;default:'[]'" json:"actions"` // JSON 数组，存储 action key 列表
}
