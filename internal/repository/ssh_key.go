package repository

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
)

// SSHKeyRepository SSH Key 数据访问层
type SSHKeyRepository struct {
	db *gorm.DB
}

// NewSSHKeyRepository 创建 SSHKeyRepository
func NewSSHKeyRepository(db *gorm.DB) *SSHKeyRepository {
	return &SSHKeyRepository{db: db}
}

// GetByID 根据 ID 查询
func (r *SSHKeyRepository) GetByID(id uint) (*model.SSHKey, error) {
	var key model.SSHKey
	if err := r.db.First(&key, id).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

// List 分页查询
func (r *SSHKeyRepository) List(keyword string, page, pageSize int) ([]model.SSHKey, int64, error) {
	var keys []model.SSHKey
	var total int64

	query := r.db.Model(&model.SSHKey{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR comment LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&keys).Error; err != nil {
		return nil, 0, err
	}

	return keys, total, nil
}

// Create 创建 SSH Key
func (r *SSHKeyRepository) Create(key *model.SSHKey) error {
	return r.db.Create(key).Error
}

// Delete 删除 SSH Key
func (r *SSHKeyRepository) Delete(id uint) error {
	return r.db.Delete(&model.SSHKey{}, id).Error
}

// CreateBinding 创建绑定关系
func (r *SSHKeyRepository) CreateBinding(binding *model.SSHKeyBinding) error {
	return r.db.Create(binding).Error
}

// ListBindingsByKeyID 查询某个 Key 的所有绑定
func (r *SSHKeyRepository) ListBindingsByKeyID(keyID uint) ([]model.SSHKeyBinding, error) {
	var bindings []model.SSHKeyBinding
	if err := r.db.Preload("Asset").Where("ssh_key_id = ?", keyID).Find(&bindings).Error; err != nil {
		return nil, err
	}
	return bindings, nil
}

// UpdateBindingStatus 更新绑定状态
func (r *SSHKeyRepository) UpdateBindingStatus(id uint, status string, updates map[string]interface{}) error {
	if updates == nil {
		updates = map[string]interface{}{}
	}
	updates["status"] = status
	return r.db.Model(&model.SSHKeyBinding{}).Where("id = ?", id).Updates(updates).Error
}

// ListBindingsByAssetID 查询某资产的所有 SSH Key 绑定
func (r *SSHKeyRepository) ListBindingsByAssetID(assetID uint) ([]model.SSHKeyBinding, error) {
	var bindings []model.SSHKeyBinding
	if err := r.db.Preload("Asset").Where("asset_id = ?", assetID).Find(&bindings).Error; err != nil {
		return nil, err
	}
	return bindings, nil
}

// DeleteBindingsByKeyID 删除某个 Key 的所有绑定
func (r *SSHKeyRepository) DeleteBindingsByKeyID(keyID uint) error {
	return r.db.Where("ssh_key_id = ?", keyID).Delete(&model.SSHKeyBinding{}).Error
}

// DeleteBinding 删除单条绑定
func (r *SSHKeyRepository) DeleteBinding(id uint) error {
	return r.db.Delete(&model.SSHKeyBinding{}, id).Error
}

// GetBindingByKeyAssetUser 查询指定密钥、资产、用户的绑定
func (r *SSHKeyRepository) GetBindingByKeyAssetUser(keyID, assetID uint, username string) (*model.SSHKeyBinding, error) {
	var binding model.SSHKeyBinding
	err := r.db.Where("ssh_key_id = ? AND asset_id = ? AND username = ?", keyID, assetID, username).First(&binding).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

// ListBindingsByUsername 查询某用户名的所有绑定（跨所有密钥）
func (r *SSHKeyRepository) ListBindingsByUsername(username string) ([]model.SSHKeyBinding, error) {
	var bindings []model.SSHKeyBinding
	if err := r.db.Preload("Asset").Where("username = ? AND status = ?", username, "deployed").Find(&bindings).Error; err != nil {
		return nil, err
	}
	return bindings, nil
}

// ListDeployedBindingsByKeyID 查询某个 Key 的已部署绑定
func (r *SSHKeyRepository) ListDeployedBindingsByKeyID(keyID uint) ([]model.SSHKeyBinding, error) {
	var bindings []model.SSHKeyBinding
	if err := r.db.Preload("Asset").Where("ssh_key_id = ? AND status = ?", keyID, "deployed").Find(&bindings).Error; err != nil {
		return nil, err
	}
	return bindings, nil
}

// CreateDeployLogs 批量创建部署日志
func (r *SSHKeyRepository) CreateDeployLogs(logs []model.SSHKeyDeployLog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.Create(&logs).Error
}

// ListDeployLogsByKeyID 查询某个 Key 的部署日志（分页）
func (r *SSHKeyRepository) ListDeployLogsByKeyID(keyID uint, page, pageSize int) ([]model.SSHKeyDeployLog, int64, error) {
	var logs []model.SSHKeyDeployLog
	var total int64

	query := r.db.Model(&model.SSHKeyDeployLog{}).Where("ssh_key_id = ?", keyID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
