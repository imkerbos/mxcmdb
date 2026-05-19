package repository

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
)

// PlaybookRepository 剧本数据访问层
type PlaybookRepository struct {
	db *gorm.DB
}

// NewPlaybookRepository 创建 PlaybookRepository
func NewPlaybookRepository(db *gorm.DB) *PlaybookRepository {
	return &PlaybookRepository{db: db}
}

// List 查询剧本列表
func (r *PlaybookRepository) List() ([]model.Playbook, error) {
	var playbooks []model.Playbook
	if err := r.db.Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).Order("is_builtin DESC, created_at DESC").Find(&playbooks).Error; err != nil {
		return nil, err
	}
	return playbooks, nil
}

// GetByID 根据 ID 查询（含步骤）
func (r *PlaybookRepository) GetByID(id uint) (*model.Playbook, error) {
	var pb model.Playbook
	if err := r.db.Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).First(&pb, id).Error; err != nil {
		return nil, err
	}
	return &pb, nil
}

// Create 创建剧本
func (r *PlaybookRepository) Create(pb *model.Playbook) error {
	return r.db.Create(pb).Error
}

// Update 更新剧本基本信息
func (r *PlaybookRepository) Update(pb *model.Playbook) error {
	return r.db.Save(pb).Error
}

// Delete 删除剧本及其步骤
func (r *PlaybookRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("playbook_id = ?", id).Delete(&model.PlaybookStep{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Playbook{}, id).Error
	})
}

// ReplaceSteps 替换剧本步骤（删除旧的，创建新的）
func (r *PlaybookRepository) ReplaceSteps(playbookID uint, steps []model.PlaybookStep) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("playbook_id = ?", playbookID).Delete(&model.PlaybookStep{}).Error; err != nil {
			return err
		}
		if len(steps) > 0 {
			return tx.Create(&steps).Error
		}
		return nil
	})
}

// CreateExecution 创建执行记录
func (r *PlaybookRepository) CreateExecution(exec *model.PlaybookExecution) error {
	return r.db.Create(exec).Error
}

// UpdateExecution 更新执行记录
func (r *PlaybookRepository) UpdateExecution(exec *model.PlaybookExecution) error {
	return r.db.Save(exec).Error
}

// CreateExecutionResult 创建单资产执行结果
func (r *PlaybookRepository) CreateExecutionResult(result *model.PlaybookExecutionResult) error {
	return r.db.Create(result).Error
}

// ListExecutions 查询执行历史
func (r *PlaybookRepository) ListExecutions(page, pageSize int) ([]model.PlaybookExecution, int64, error) {
	var total int64
	var execs []model.PlaybookExecution

	r.db.Model(&model.PlaybookExecution{}).Count(&total)

	offset := (page - 1) * pageSize
	if err := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&execs).Error; err != nil {
		return nil, 0, err
	}
	return execs, total, nil
}

// GetExecutionResults 查询某次执行的所有资产结果
func (r *PlaybookRepository) GetExecutionResults(executionID uint) ([]model.PlaybookExecutionResult, error) {
	var results []model.PlaybookExecutionResult
	if err := r.db.Where("execution_id = ?", executionID).Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}
