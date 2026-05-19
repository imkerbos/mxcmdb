package model

import "time"

// FileTask 文件分发任务
type FileTask struct {
	BaseModel
	Name       string     `gorm:"size:128;not null" json:"name"`
	FileName   string     `gorm:"size:256;not null" json:"file_name"`
	FileSize   int64      `json:"file_size"`
	RemotePath string     `gorm:"size:512;not null" json:"remote_path"`
	FileMode   string     `gorm:"size:10;default:0644" json:"file_mode"`
	Status     string     `gorm:"size:32;index;not null" json:"status"` // pending / running / completed / failed
	CreatedBy  uint       `gorm:"index" json:"created_by"`
	TotalCount int        `json:"total_count"`
	SuccessCount int      `json:"success_count"`
	FailCount  int        `json:"fail_count"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
}

// FileTaskResult 文件分发结果
type FileTaskResult struct {
	BaseModel
	FileTaskID uint   `gorm:"index;not null" json:"file_task_id"`
	AssetID    uint   `gorm:"index;not null" json:"asset_id"`
	Hostname   string `gorm:"size:128" json:"hostname"`
	IP         string `gorm:"size:45" json:"ip"`
	Status     string `gorm:"size:32" json:"status"` // pending / success / failed
	ErrorMsg   string `gorm:"type:text" json:"error_msg"`
	Duration   int64  `json:"duration"` // 毫秒
}
