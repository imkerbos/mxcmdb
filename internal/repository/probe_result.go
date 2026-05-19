package repository

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
)

// ProbeResultRepository 探针结果数据访问层
type ProbeResultRepository struct {
	db *gorm.DB
}

// NewProbeResultRepository 创建 ProbeResultRepository
func NewProbeResultRepository(db *gorm.DB) *ProbeResultRepository {
	return &ProbeResultRepository{db: db}
}

// Create 创建探针结果
func (r *ProbeResultRepository) Create(result *model.ProbeResult) error {
	return r.db.Create(result).Error
}

// GetLatestByAssetID 获取某资产最新的探针结果
func (r *ProbeResultRepository) GetLatestByAssetID(assetID uint) (*model.ProbeResult, error) {
	var result model.ProbeResult
	if err := r.db.Where("asset_id = ?", assetID).Order("collected_at DESC").First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

// ListByAssetID 查询某资产的历史探针结果
func (r *ProbeResultRepository) ListByAssetID(assetID uint, limit int) ([]model.ProbeResult, error) {
	var results []model.ProbeResult
	if err := r.db.Where("asset_id = ?", assetID).Order("collected_at DESC").Limit(limit).Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

// GetLatestByAssetIDs 批量获取多个资产的最新探针结果
func (r *ProbeResultRepository) GetLatestByAssetIDs(assetIDs []uint) ([]model.ProbeResult, error) {
	if len(assetIDs) == 0 {
		return nil, nil
	}
	// 子查询：每个 asset_id 的最新 collected_at
	var results []model.ProbeResult
	subQuery := r.db.Model(&model.ProbeResult{}).
		Select("asset_id, MAX(collected_at) as max_collected").
		Where("asset_id IN ? AND status = ?", assetIDs, "success").
		Group("asset_id")

	if err := r.db.Where("(asset_id, collected_at) IN (?)", subQuery).
		Find(&results).Error; err != nil {
		// 回退到逐个查询（兼容不支持 tuple IN 的 DB）
		results = nil
		for _, id := range assetIDs {
			var pr model.ProbeResult
			if err := r.db.Where("asset_id = ? AND status = ?", id, "success").
				Order("collected_at DESC").First(&pr).Error; err == nil {
				results = append(results, pr)
			}
		}
	}
	return results, nil
}
