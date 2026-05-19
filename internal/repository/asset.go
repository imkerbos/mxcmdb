package repository

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
)

// AssetRepository 资产数据访问层
type AssetRepository struct {
	db *gorm.DB
}

// NewAssetRepository 创建 AssetRepository
func NewAssetRepository(db *gorm.DB) *AssetRepository {
	return &AssetRepository{db: db}
}

// GetByID 根据 ID 查询（含标签）
func (r *AssetRepository) GetByID(id uint) (*model.Asset, error) {
	var asset model.Asset
	if err := r.db.Preload("Tags").First(&asset, id).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

// GetByIDs 根据 ID 列表批量查询
func (r *AssetRepository) GetByIDs(ids []uint) ([]model.Asset, error) {
	var assets []model.Asset
	if len(ids) == 0 {
		return assets, nil
	}
	if err := r.db.Where("id IN ?", ids).Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

// GetByIP 根据 IP 查询
func (r *AssetRepository) GetByIP(ip string) (*model.Asset, error) {
	var asset model.Asset
	if err := r.db.Where("ip = ?", ip).First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

// GetByInstanceID 根据云实例 ID 查询
func (r *AssetRepository) GetByInstanceID(instanceID string) (*model.Asset, error) {
	var asset model.Asset
	if err := r.db.Where("instance_id = ?", instanceID).First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

// List 分页查询
func (r *AssetRepository) List(keyword, assetType, source, status, env, dept string, projectID uint, owner string, page, pageSize int) ([]model.Asset, int64, error) {
	var assets []model.Asset
	var total int64

	query := r.db.Model(&model.Asset{})

	if keyword != "" {
		query = query.Where("hostname LIKE ? OR ip LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if assetType != "" {
		query = query.Where("type = ?", assetType)
	}
	if source != "" {
		query = query.Where("source = ?", source)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if env != "" {
		query = query.Where("environment = ?", env)
	}
	if dept != "" {
		query = query.Where("department = ?", dept)
	}
	if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	}
	if owner != "" {
		query = query.Where("owner = ?", owner)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Tags").Order("id DESC").Offset(offset).Limit(pageSize).Find(&assets).Error; err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

// Create 创建资产
func (r *AssetRepository) Create(asset *model.Asset) error {
	return r.db.Create(asset).Error
}

// Update 更新资产
func (r *AssetRepository) Update(asset *model.Asset) error {
	return r.db.Save(asset).Error
}

// Delete 软删除资产
func (r *AssetRepository) Delete(id uint) error {
	return r.db.Delete(&model.Asset{}, id).Error
}

// CountByType 按类型统计
func (r *AssetRepository) CountByType(assetType string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Asset{}).Where("type = ?", assetType).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountBySource 按来源统计
func (r *AssetRepository) CountBySource(source string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Asset{}).Where("source = ?", source).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountByStatus 按状态统计
func (r *AssetRepository) CountByStatus(status string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Asset{}).Where("status = ?", status).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountProbed 统计已探测数量
func (r *AssetRepository) CountProbed() (int64, error) {
	var count int64
	if err := r.db.Model(&model.Asset{}).Where("probe_last_at IS NOT NULL").Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// DeleteTagsByAssetID 删除资产的所有标签
func (r *AssetRepository) DeleteTagsByAssetID(assetID uint) error {
	return r.db.Where("asset_id = ?", assetID).Delete(&model.AssetTag{}).Error
}

// CreateTags 批量创建标签
func (r *AssetRepository) CreateTags(tags []model.AssetTag) error {
	if len(tags) == 0 {
		return nil
	}
	return r.db.Create(&tags).Error
}

// GroupByEnvironment 按环境统计
func (r *AssetRepository) GroupByEnvironment() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	if err := r.db.Model(&model.Asset{}).
		Select("COALESCE(NULLIF(environment,''), 'unknown') as label, COUNT(*) as value").
		Group("label").Order("value DESC").
		Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

// GroupByType 按类型统计
func (r *AssetRepository) GroupByType() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	if err := r.db.Model(&model.Asset{}).
		Select("COALESCE(NULLIF(type,''), 'unknown') as label, COUNT(*) as value").
		Group("label").Order("value DESC").
		Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

// ListAll 获取所有资产（用于选择器）
func (r *AssetRepository) ListAll() ([]model.Asset, error) {
	var assets []model.Asset
	if err := r.db.Select("id, hostname, ip, type, status, environment, project_id").Order("hostname").Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

// ListBySource 按来源查询资产（含 SSH 连接信息）
func (r *AssetRepository) ListBySource(source string) ([]model.Asset, error) {
	var assets []model.Asset
	if err := r.db.Where("source = ?", source).Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}
