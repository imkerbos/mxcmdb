package model

// Permission 权限定义
type Permission struct {
	BaseModel
	Code        string `gorm:"uniqueIndex;size:64;not null" json:"code"`    // 权限标识: asset:read, asset:write, sshkey:deploy
	Name        string `gorm:"size:128;not null" json:"name"`               // 权限名称
	Module      string `gorm:"size:64;not null;index" json:"module"`        // 所属模块: asset, sshkey, probe, task, user, audit, system
	Description string `gorm:"size:256" json:"description"`                 // 权限描述
}

// RolePermission 角色-权限映射
type RolePermission struct {
	BaseModel
	Role           string `gorm:"size:32;not null;index:idx_role_perm,unique" json:"role"`                // 角色名
	PermissionCode string `gorm:"size:64;not null;index:idx_role_perm,unique" json:"permission_code"`     // 权限标识
}
