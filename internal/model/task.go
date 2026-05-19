package model

import "time"

// Task 批量任务
type Task struct {
	BaseModel
	Name         string     `gorm:"size:128;not null" json:"name"`
	Type         string     `gorm:"size:32;index;not null" json:"type"`   // command
	Command      string     `gorm:"type:text;not null" json:"command"`
	Status       string     `gorm:"size:32;index;not null" json:"status"` // pending / running / completed / failed
	CreatedBy    uint       `gorm:"index" json:"created_by"`
	CreatedByName string    `gorm:"-" json:"created_by_name,omitempty"`
	TotalCount   int        `json:"total_count"`
	SuccessCount int        `json:"success_count"`
	FailCount    int        `json:"fail_count"`
	StartedAt    *time.Time `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
}

// TaskResult 任务执行结果
type TaskResult struct {
	BaseModel
	TaskID   uint   `gorm:"index;not null" json:"task_id"`
	AssetID  uint   `gorm:"index;not null" json:"asset_id"`
	Hostname string `gorm:"size:128" json:"hostname"`
	IP       string `gorm:"size:45" json:"ip"`
	Status   string `gorm:"size:32" json:"status"` // pending / running / success / failed / timeout
	Stdout   string `gorm:"type:text" json:"stdout"`
	Stderr   string `gorm:"type:text" json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Duration int64  `json:"duration"` // 毫秒
}
