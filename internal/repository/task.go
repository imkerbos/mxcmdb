package repository

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
)

// TaskRepository 任务数据访问
type TaskRepository struct {
	db *gorm.DB
}

// NewTaskRepository 创建 TaskRepository
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create 创建任务
func (r *TaskRepository) Create(task *model.Task) error {
	return r.db.Create(task).Error
}

// GetByID 按 ID 查询
func (r *TaskRepository) GetByID(id uint) (*model.Task, error) {
	var task model.Task
	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// Update 更新任务
func (r *TaskRepository) Update(task *model.Task) error {
	return r.db.Save(task).Error
}

// List 任务列表
func (r *TaskRepository) List(status string, page, pageSize int) ([]model.Task, int64, error) {
	var tasks []model.Task
	var total int64

	query := r.db.Model(&model.Task{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

// CreateResult 创建执行结果
func (r *TaskRepository) CreateResult(result *model.TaskResult) error {
	return r.db.Create(result).Error
}

// CreateResults 批量创建执行结果
func (r *TaskRepository) CreateResults(results []*model.TaskResult) error {
	if len(results) == 0 {
		return nil
	}
	return r.db.CreateInBatches(results, 100).Error
}

// UpdateResult 更新执行结果
func (r *TaskRepository) UpdateResult(result *model.TaskResult) error {
	return r.db.Save(result).Error
}

// ListResults 查询任务结果
func (r *TaskRepository) ListResults(taskID uint) ([]model.TaskResult, error) {
	var results []model.TaskResult
	if err := r.db.Where("task_id = ?", taskID).Order("id ASC").Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}
