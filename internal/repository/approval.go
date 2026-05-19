package repository

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
)

// ApprovalRepository 审批数据访问层
type ApprovalRepository struct {
	db *gorm.DB
}

// NewApprovalRepository 创建 ApprovalRepository
func NewApprovalRepository(db *gorm.DB) *ApprovalRepository {
	return &ApprovalRepository{db: db}
}

// Create 创建审批工单
func (r *ApprovalRepository) Create(a *model.Approval) error {
	return r.db.Create(a).Error
}

// GetByID 根据 ID 查询
func (r *ApprovalRepository) GetByID(id uint) (*model.Approval, error) {
	var a model.Approval
	if err := r.db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// List 分页查询
func (r *ApprovalRepository) List(status, approvalType string, page, pageSize int) ([]model.Approval, int64, error) {
	var items []model.Approval
	var total int64

	query := r.db.Model(&model.Approval{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if approvalType != "" {
		query = query.Where("type = ?", approvalType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// ListByRequester 查询用户发起的审批
func (r *ApprovalRepository) ListByRequester(userID uint, page, pageSize int) ([]model.Approval, int64, error) {
	var items []model.Approval
	var total int64

	query := r.db.Model(&model.Approval{}).Where("requested_by = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// Update 更新审批工单
func (r *ApprovalRepository) Update(a *model.Approval) error {
	return r.db.Save(a).Error
}

// CountPending 统计待审批数量
func (r *ApprovalRepository) CountPending() (int64, error) {
	var count int64
	err := r.db.Model(&model.Approval{}).Where("status = ?", "pending").Count(&count).Error
	return count, err
}
