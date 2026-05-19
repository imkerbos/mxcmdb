package model

import "time"

// Playbook 运维剧本
type Playbook struct {
	BaseModel
	Name        string         `gorm:"size:128;not null" json:"name"`
	Description string         `gorm:"size:512" json:"description"`
	Category    string         `gorm:"size:64;not null;index" json:"category"` // builtin / custom
	IsBuiltin   bool           `gorm:"default:false" json:"is_builtin"`
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	Steps       []PlaybookStep `gorm:"foreignKey:PlaybookID" json:"steps"`
	CreatedBy   uint           `json:"created_by"`
}

// PlaybookStep 剧本步骤
type PlaybookStep struct {
	BaseModel
	PlaybookID   uint   `gorm:"index;not null" json:"playbook_id"`
	SortOrder    int    `gorm:"not null;default:0" json:"sort_order"`
	Name         string `gorm:"size:128;not null" json:"name"`
	Script       string `gorm:"type:text;not null" json:"script"`
	Timeout      int    `gorm:"default:60" json:"timeout"`       // 秒
	OnError      string `gorm:"size:32;default:stop" json:"on_error"` // stop / continue
}

// PlaybookExecution 剧本执行记录
type PlaybookExecution struct {
	BaseModel
	PlaybookID   uint       `gorm:"not null;index" json:"playbook_id"`  // 主剧本 ID（首个）
	PlaybookIDs  string     `gorm:"size:256" json:"playbook_ids"`       // 逗号分隔的剧本 ID
	PlaybookName string     `gorm:"size:256" json:"playbook_name"`      // 逗号分隔的剧本名称
	Status       string     `gorm:"size:32;default:running" json:"status"` // running / success / partial / failed
	TotalAssets  int        `json:"total_assets"`
	Success      int        `json:"success"`
	Failed       int        `json:"failed"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	CreatedBy    uint       `json:"created_by"`
}

// PlaybookExecutionResult 单资产执行结果
type PlaybookExecutionResult struct {
	BaseModel
	ExecutionID uint   `gorm:"index;not null" json:"execution_id"`
	AssetID     uint   `gorm:"index;not null" json:"asset_id"`
	Hostname    string `gorm:"size:128" json:"hostname"`
	IP          string `gorm:"size:45" json:"ip"`
	Status      string `gorm:"size:32" json:"status"` // success / partial / failed
	StepResults string `gorm:"type:text" json:"step_results"` // JSON 数组
}
