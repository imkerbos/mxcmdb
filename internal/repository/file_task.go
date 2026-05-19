package repository

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
)

// FileTaskRepository 文件任务数据访问
type FileTaskRepository struct {
	db *gorm.DB
}

// NewFileTaskRepository 创建 FileTaskRepository
func NewFileTaskRepository(db *gorm.DB) *FileTaskRepository {
	return &FileTaskRepository{db: db}
}

// Create 创建文件任务
func (r *FileTaskRepository) Create(task *model.FileTask) error {
	return r.db.Create(task).Error
}

// GetByID 按 ID 查询
func (r *FileTaskRepository) GetByID(id uint) (*model.FileTask, error) {
	var task model.FileTask
	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// Update 更新文件任务
func (r *FileTaskRepository) Update(task *model.FileTask) error {
	return r.db.Save(task).Error
}

// List 文件任务列表
func (r *FileTaskRepository) List(page, pageSize int) ([]model.FileTask, int64, error) {
	var tasks []model.FileTask
	var total int64

	if err := r.db.Model(&model.FileTask{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

// CreateResult 创建结果
func (r *FileTaskRepository) CreateResult(result *model.FileTaskResult) error {
	return r.db.Create(result).Error
}

// CreateResults 批量创建结果
func (r *FileTaskRepository) CreateResults(results []*model.FileTaskResult) error {
	if len(results) == 0 {
		return nil
	}
	return r.db.CreateInBatches(results, 100).Error
}

// UpdateResult 更新结果
func (r *FileTaskRepository) UpdateResult(result *model.FileTaskResult) error {
	return r.db.Save(result).Error
}

// ListResults 查询结果
func (r *FileTaskRepository) ListResults(fileTaskID uint) ([]model.FileTaskResult, error) {
	var results []model.FileTaskResult
	if err := r.db.Where("file_task_id = ?", fileTaskID).Order("id ASC").Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}
