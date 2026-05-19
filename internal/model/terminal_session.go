package model

import "time"

// TerminalSession Web Terminal 会话
type TerminalSession struct {
	BaseModel
	AssetID    uint       `gorm:"index;not null" json:"asset_id"`
	Asset      Asset      `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
	UserID     uint       `gorm:"index;not null" json:"user_id"`
	Username   string     `gorm:"size:64" json:"username"`
	Status     string     `gorm:"size:32;index;index:idx_terminal_session_status_id,priority:1" json:"status"` // connected / disconnected
	ClientIP   string     `gorm:"size:45" json:"client_ip"`
	Recording  string     `gorm:"type:text" json:"-"`          // asciicast v2 格式
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
}
