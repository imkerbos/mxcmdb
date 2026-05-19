package repository

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SystemConfigRepository 系统配置数据访问层
type SystemConfigRepository struct {
	db *gorm.DB
}

// NewSystemConfigRepository 创建 SystemConfigRepository
func NewSystemConfigRepository(db *gorm.DB) *SystemConfigRepository {
	return &SystemConfigRepository{db: db}
}

// GetByKey 根据 key 查询
func (r *SystemConfigRepository) GetByKey(key string) (*model.SystemConfig, error) {
	var cfg model.SystemConfig
	if err := r.db.Where("\"key\" = ?", key).First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetByCategory 根据分类查询
func (r *SystemConfigRepository) GetByCategory(category string) ([]model.SystemConfig, error) {
	var configs []model.SystemConfig
	if err := r.db.Where("category = ?", category).Order("key").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// List 查询所有配置
func (r *SystemConfigRepository) List() ([]model.SystemConfig, error) {
	var configs []model.SystemConfig
	if err := r.db.Order("category, key").Find(&configs).Error; err != nil {
		return nil, err
	}
	return configs, nil
}

// Upsert 创建或更新配置
func (r *SystemConfigRepository) Upsert(cfg *model.SystemConfig) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(cfg).Error
}

// Delete 删除配置
func (r *SystemConfigRepository) Delete(key string) error {
	return r.db.Where("\"key\" = ?", key).Delete(&model.SystemConfig{}).Error
}
