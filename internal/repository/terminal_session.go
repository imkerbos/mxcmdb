package repository

import (
	"time"

	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
)

// TerminalSessionRepository 终端会话数据访问
type TerminalSessionRepository struct {
	db *gorm.DB
}

// NewTerminalSessionRepository 创建 TerminalSessionRepository
func NewTerminalSessionRepository(db *gorm.DB) *TerminalSessionRepository {
	return &TerminalSessionRepository{db: db}
}

// CleanupStaleSessions 清理僵尸会话（进程重启后残留的 connected 记录）
func (r *TerminalSessionRepository) CleanupStaleSessions() (int64, error) {
	result := r.db.Model(&model.TerminalSession{}).
		Where("status = ?", "connected").
		Updates(map[string]any{
			"status":      "disconnected",
			"finished_at": time.Now(),
		})
	return result.RowsAffected, result.Error
}

// Create 创建会话
func (r *TerminalSessionRepository) Create(session *model.TerminalSession) error {
	return r.db.Create(session).Error
}

// Update 更新会话
func (r *TerminalSessionRepository) Update(session *model.TerminalSession) error {
	return r.db.Save(session).Error
}

// GetByID 按 ID 查询
func (r *TerminalSessionRepository) GetByID(id uint) (*model.TerminalSession, error) {
	var session model.TerminalSession
	if err := r.db.Preload("Asset").First(&session, id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// ListByAssetID 查询某资产的终端会话
func (r *TerminalSessionRepository) ListByAssetID(assetID uint, limit int) ([]model.TerminalSession, error) {
	var sessions []model.TerminalSession
	if err := r.db.Where("asset_id = ?", assetID).Order("id DESC").Limit(limit).Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

// List 会话列表，支持 status 过滤
func (r *TerminalSessionRepository) List(page, pageSize int, status string) ([]model.TerminalSession, int64, error) {
	var sessions []model.TerminalSession
	var total int64

	query := r.db.Model(&model.TerminalSession{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Asset").Order("id DESC").Offset(offset).Limit(pageSize).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}
