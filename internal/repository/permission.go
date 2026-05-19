package repository

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
)

// PermissionRepository 权限数据访问层
type PermissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository 创建 PermissionRepository
func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// ListAll 查询所有权限定义
func (r *PermissionRepository) ListAll() ([]model.Permission, error) {
	var perms []model.Permission
	if err := r.db.Order("module, code").Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// GetByCode 根据 code 查询权限
func (r *PermissionRepository) GetByCode(code string) (*model.Permission, error) {
	var perm model.Permission
	if err := r.db.Where("code = ?", code).First(&perm).Error; err != nil {
		return nil, err
	}
	return &perm, nil
}

// CreateIfNotExists 创建权限（不存在时）
func (r *PermissionRepository) CreateIfNotExists(perm *model.Permission) error {
	return r.db.Where("code = ?", perm.Code).FirstOrCreate(perm).Error
}

// ListRolePermissions 查询某角色的所有权限
func (r *PermissionRepository) ListRolePermissions(role string) ([]string, error) {
	var codes []string
	err := r.db.Model(&model.RolePermission{}).Where("role = ?", role).Pluck("permission_code", &codes).Error
	return codes, err
}

// SetRolePermissions 设置角色权限（全量替换）
func (r *PermissionRepository) SetRolePermissions(role string, codes []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧权限
		if err := tx.Where("role = ?", role).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		// 插入新权限
		for _, code := range codes {
			rp := model.RolePermission{Role: role, PermissionCode: code}
			if err := tx.Create(&rp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ListAllRolePermissions 查询所有角色的权限映射
func (r *PermissionRepository) ListAllRolePermissions() ([]model.RolePermission, error) {
	var rps []model.RolePermission
	if err := r.db.Find(&rps).Error; err != nil {
		return nil, err
	}
	return rps, nil
}
