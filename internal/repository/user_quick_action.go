package repository

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserQuickActionRepository 用户快捷操作数据访问
type UserQuickActionRepository struct {
	db *gorm.DB
}

// NewUserQuickActionRepository 创建 UserQuickActionRepository
func NewUserQuickActionRepository(db *gorm.DB) *UserQuickActionRepository {
	return &UserQuickActionRepository{db: db}
}

// GetByUserID 根据用户 ID 获取快捷操作配置
func (r *UserQuickActionRepository) GetByUserID(userID uint) (*model.UserQuickAction, error) {
	var record model.UserQuickAction
	err := r.db.Where("user_id = ?", userID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// Upsert 创建或更新用户快捷操作配置
func (r *UserQuickActionRepository) Upsert(userID uint, actions string) error {
	record := model.UserQuickAction{
		UserID:  userID,
		Actions: actions,
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"actions", "updated_at"}),
	}).Create(&record).Error
}
