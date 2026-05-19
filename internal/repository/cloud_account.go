package repository

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
)

// CloudAccountRepository 云账号数据访问层
type CloudAccountRepository struct {
	db *gorm.DB
}

// NewCloudAccountRepository 创建 CloudAccountRepository
func NewCloudAccountRepository(db *gorm.DB) *CloudAccountRepository {
	return &CloudAccountRepository{db: db}
}

// GetByID 根据 ID 查询
func (r *CloudAccountRepository) GetByID(id uint) (*model.CloudAccount, error) {
	var account model.CloudAccount
	if err := r.db.First(&account, id).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// List 分页查询
func (r *CloudAccountRepository) List(keyword, provider string, status *int, page, pageSize int) ([]model.CloudAccount, int64, error) {
	var accounts []model.CloudAccount
	var total int64

	query := r.db.Model(&model.CloudAccount{})

	if keyword != "" {
		query = query.Where("name LIKE ? OR access_key_id LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if provider != "" {
		query = query.Where("provider = ?", provider)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&accounts).Error; err != nil {
		return nil, 0, err
	}

	return accounts, total, nil
}

// Create 创建云账号
func (r *CloudAccountRepository) Create(account *model.CloudAccount) error {
	return r.db.Create(account).Error
}

// Update 更新云账号
func (r *CloudAccountRepository) Update(account *model.CloudAccount) error {
	return r.db.Save(account).Error
}

// Delete 软删除云账号
func (r *CloudAccountRepository) Delete(id uint) error {
	return r.db.Delete(&model.CloudAccount{}, id).Error
}

// ListActive 获取所有启用的云账号
func (r *CloudAccountRepository) ListActive() ([]model.CloudAccount, error) {
	var accounts []model.CloudAccount
	if err := r.db.Where("status = 1").Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}
