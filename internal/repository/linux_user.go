package repository

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
)

// LinuxUserRepository Linux 用户数据访问
type LinuxUserRepository struct {
	db *gorm.DB
}

// NewLinuxUserRepository 创建 LinuxUserRepository
func NewLinuxUserRepository(db *gorm.DB) *LinuxUserRepository {
	return &LinuxUserRepository{db: db}
}

// Create 创建记录
func (r *LinuxUserRepository) Create(user *model.LinuxUser) error {
	return r.db.Create(user).Error
}

// Update 更新记录
func (r *LinuxUserRepository) Update(user *model.LinuxUser) error {
	return r.db.Save(user).Error
}

// Delete 删除记录
func (r *LinuxUserRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinuxUser{}, id).Error
}

// GetByID 按 ID 查询
func (r *LinuxUserRepository) GetByID(id uint) (*model.LinuxUser, error) {
	var user model.LinuxUser
	if err := r.db.Preload("Asset").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// List 列表查询
func (r *LinuxUserRepository) List(username string, page, pageSize int) ([]model.LinuxUser, int64, error) {
	var users []model.LinuxUser
	var total int64

	query := r.db.Model(&model.LinuxUser{}).Preload("Asset")
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// ListByAssetID 查询某资产的所有 Linux 用户
func (r *LinuxUserRepository) ListByAssetID(assetID uint) ([]model.LinuxUser, error) {
	var users []model.LinuxUser
	if err := r.db.Where("asset_id = ? AND status != ?", assetID, "deleted").Order("id DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// FindByUsernameAndAsset 按用户名和资产查找
func (r *LinuxUserRepository) FindByUsernameAndAsset(username string, assetID uint) (*model.LinuxUser, error) {
	var user model.LinuxUser
	if err := r.db.Where("username = ? AND asset_id = ?", username, assetID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
