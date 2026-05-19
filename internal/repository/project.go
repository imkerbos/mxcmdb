package repository

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
)

// ProjectRepository 项目数据访问层
type ProjectRepository struct {
	db *gorm.DB
}

// NewProjectRepository 创建 ProjectRepository
func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// GetByID 根据 ID 查询
func (r *ProjectRepository) GetByID(id uint) (*model.Project, error) {
	var project model.Project
	if err := r.db.First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// GetByCode 根据 Code 查询
func (r *ProjectRepository) GetByCode(code string) (*model.Project, error) {
	var project model.Project
	if err := r.db.Where("code = ?", code).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// GetByName 根据名称查询
func (r *ProjectRepository) GetByName(name string) (*model.Project, error) {
	var project model.Project
	if err := r.db.Where("name = ?", name).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// List 分页查询
func (r *ProjectRepository) List(keyword, status string, page, pageSize int) ([]model.Project, int64, error) {
	var projects []model.Project
	var total int64

	query := r.db.Model(&model.Project{})

	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// ListAll 获取所有项目（用于选择器）
func (r *ProjectRepository) ListAll() ([]model.Project, error) {
	var projects []model.Project
	if err := r.db.Where("status = ?", "active").Select("id, name, code").Order("name").Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

// Create 创建项目
func (r *ProjectRepository) Create(project *model.Project) error {
	return r.db.Create(project).Error
}

// Update 更新项目
func (r *ProjectRepository) Update(project *model.Project) error {
	return r.db.Save(project).Error
}

// Delete 软删除项目
func (r *ProjectRepository) Delete(id uint) error {
	return r.db.Delete(&model.Project{}, id).Error
}

// Count 统计项目总数
func (r *ProjectRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&model.Project{}).Where("status = ?", "active").Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountAssets 统计项目关联的资产数
func (r *ProjectRepository) CountAssets(projectID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Asset{}).Where("project_id = ?", projectID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ListAssetsByProjectID 查询项目下所有资产
func (r *ProjectRepository) ListAssetsByProjectID(projectID uint, page, pageSize int) ([]model.Asset, int64, error) {
	var assets []model.Asset
	var total int64

	query := r.db.Model(&model.Asset{}).Where("project_id = ?", projectID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Tags").Order("id DESC").Offset(offset).Limit(pageSize).Find(&assets).Error; err != nil {
		return nil, 0, err
	}
	return assets, total, nil
}

// ListAssetIDsByProjectID 查询项目下所有资产 ID
func (r *ProjectRepository) ListAssetIDsByProjectID(projectID uint) ([]uint, error) {
	var ids []uint
	if err := r.db.Model(&model.Asset{}).Where("project_id = ?", projectID).Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
