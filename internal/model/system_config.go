package model

// SystemConfig 系统配置模型（业务配置存数据库）
type SystemConfig struct {
	BaseModel
	Key         string `gorm:"uniqueIndex;size:128;not null" json:"key"`
	Value       string `gorm:"type:text;not null" json:"value"`
	Category    string `gorm:"size:64;not null;index" json:"category"`
	Description string `gorm:"size:256" json:"description"`
	Type        string `gorm:"size:32;default:string" json:"type"` // string / int / bool / json
}
