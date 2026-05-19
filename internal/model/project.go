package model

// Project 项目模型
type Project struct {
	BaseModel
	Name        string `gorm:"size:64;not null;uniqueIndex" json:"name"`
	Code        string `gorm:"size:32;uniqueIndex" json:"code"`
	Description string `gorm:"size:256" json:"description"`
	Owner       string `gorm:"size:64" json:"owner"`
	Status      string `gorm:"size:16;default:active;index" json:"status"` // active / archived
}
