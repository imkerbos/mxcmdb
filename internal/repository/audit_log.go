package repository

import (
	"time"

	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
)

// AuditLogRepository 审计日志数据访问层
type AuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository 创建 AuditLogRepository
func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create 创建审计日志
func (r *AuditLogRepository) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

// BatchCreate 批量创建审计日志
func (r *AuditLogRepository) BatchCreate(logs []model.AuditLog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.Create(&logs).Error
}

// List 分页查询审计日志
func (r *AuditLogRepository) List(module, action, username, startDate, endDate string, page, pageSize int) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	query := r.db.Model(&model.AuditLog{})

	if module != "" {
		query = query.Where("module = ?", module)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if startDate != "" {
		t, err := time.Parse("2006-01-02", startDate)
		if err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endDate != "" {
		t, err := time.Parse("2006-01-02", endDate)
		if err == nil {
			query = query.Where("created_at < ?", t.AddDate(0, 0, 1))
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByID 根据 ID 查询
func (r *AuditLogRepository) GetByID(id uint) (*model.AuditLog, error) {
	var log model.AuditLog
	if err := r.db.First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}
