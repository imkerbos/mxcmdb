package repository

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建 UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByUsername 根据用户名查询
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID 根据 ID 查询
func (r *UserRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// Update 更新用户
func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// Count 统计用户数量
func (r *UserRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&model.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// List 分页查询用户列表
func (r *UserRepository) List(keyword, role string, status *int, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.Model(&model.User{})

	if keyword != "" {
		query = query.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
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

// Delete 软删除用户
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

// ListByRole 查询指定角色的所有活跃用户
func (r *UserRepository) ListByRole(role string) ([]model.User, error) {
	var users []model.User
	if err := r.db.Where("role = ? AND status = ?", role, 1).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// GetByEmail 根据邮箱查询
func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
