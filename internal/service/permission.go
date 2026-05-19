package service

import (
	"sync"

	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/pkg/logger"
	"github.com/imkerbos/mxcmdb/internal/repository"
)

// 默认权限定义
var defaultPermissions = []model.Permission{
	// 资产模块
	{Code: "asset:read", Name: "查看资产", Module: "asset", Description: "查看资产列表和详情"},
	{Code: "asset:write", Name: "编辑资产", Module: "asset", Description: "创建、编辑、导入资产"},
	{Code: "asset:delete", Name: "删除资产", Module: "asset", Description: "删除资产"},
	{Code: "asset:test_conn", Name: "测试连接", Module: "asset", Description: "测试资产 SSH 连接"},
	{Code: "asset:init", Name: "资产初始化", Module: "asset", Description: "执行资产初始化流程"},

	// 云账号模块
	{Code: "cloud:read", Name: "查看云账号", Module: "cloud", Description: "查看云账号列表"},
	{Code: "cloud:write", Name: "管理云账号", Module: "cloud", Description: "创建、编辑、删除云账号"},

	// SSH Key 模块
	{Code: "sshkey:read", Name: "查看密钥", Module: "sshkey", Description: "查看 SSH Key 列表和绑定"},
	{Code: "sshkey:write", Name: "管理密钥", Module: "sshkey", Description: "创建、删除 SSH Key"},
	{Code: "sshkey:deploy", Name: "部署密钥", Module: "sshkey", Description: "部署 SSH Key 到资产"},
	{Code: "sshkey:revoke", Name: "撤销密钥", Module: "sshkey", Description: "从资产撤销 SSH Key"},
	{Code: "sshkey:rotate", Name: "轮换密钥", Module: "sshkey", Description: "SSH Key 轮换"},
	{Code: "sshkey:departure", Name: "离职清理", Module: "sshkey", Description: "SSH Key 离职清理"},

	// 探针模块
	{Code: "probe:read", Name: "查看探针", Module: "probe", Description: "查看探针采集结果"},
	{Code: "probe:execute", Name: "执行探针", Module: "probe", Description: "触发探针采集"},

	// 任务模块
	{Code: "task:read", Name: "查看任务", Module: "task", Description: "查看任务列表和结果"},
	{Code: "task:execute", Name: "执行任务", Module: "task", Description: "执行批量命令"},

	// 文件分发
	{Code: "file:read", Name: "查看分发", Module: "file", Description: "查看文件分发记录"},
	{Code: "file:distribute", Name: "文件分发", Module: "file", Description: "执行文件分发"},

	// 终端
	{Code: "terminal:connect", Name: "终端连接", Module: "terminal", Description: "使用 Web Terminal 连接资产"},
	{Code: "terminal:session", Name: "查看会话", Module: "terminal", Description: "查看终端会话记录和回放"},

	// Linux 用户
	{Code: "linux_user:read", Name: "查看用户", Module: "linux_user", Description: "查看 Linux 用户列表"},
	{Code: "linux_user:write", Name: "管理用户", Module: "linux_user", Description: "创建、删除 Linux 用户"},

	// 项目
	{Code: "project:read", Name: "查看项目", Module: "project", Description: "查看项目列表和详情"},
	{Code: "project:write", Name: "管理项目", Module: "project", Description: "创建、编辑、删除项目"},

	// IPAM
	{Code: "ipam:read", Name: "查看 IPAM", Module: "ipam", Description: "查看网段和 IP 分配"},
	{Code: "ipam:write", Name: "管理 IPAM", Module: "ipam", Description: "创建网段、分配 IP"},

	// 审计
	{Code: "audit:read", Name: "查看审计", Module: "audit", Description: "查看审计日志"},

	// 审批
	{Code: "approval:read", Name: "查看审批", Module: "approval", Description: "查看审批工单"},
	{Code: "approval:review", Name: "审批操作", Module: "approval", Description: "通过或驳回审批工单"},

	// 运维剧本
	{Code: "playbook:read", Name: "查看剧本", Module: "playbook", Description: "查看运维剧本"},
	{Code: "playbook:write", Name: "管理剧本", Module: "playbook", Description: "创建、编辑、删除剧本"},
	{Code: "playbook:execute", Name: "执行剧本", Module: "playbook", Description: "执行运维剧本"},

	// 系统管理
	{Code: "user:read", Name: "查看用户", Module: "system", Description: "查看平台用户列表"},
	{Code: "user:write", Name: "管理用户", Module: "system", Description: "创建、编辑、删除平台用户"},
	{Code: "setting:read", Name: "查看配置", Module: "system", Description: "查看系统配置"},
	{Code: "setting:write", Name: "修改配置", Module: "system", Description: "修改系统配置"},
	{Code: "permission:manage", Name: "权限管理", Module: "system", Description: "管理角色权限"},
}

// 默认角色权限映射
var defaultRolePermissions = map[string][]string{
	"admin": {
		"asset:read", "asset:write", "asset:delete", "asset:test_conn", "asset:init",
		"cloud:read", "cloud:write",
		"sshkey:read", "sshkey:write", "sshkey:deploy", "sshkey:revoke", "sshkey:rotate", "sshkey:departure",
		"probe:read", "probe:execute",
		"task:read", "task:execute",
		"file:read", "file:distribute",
		"terminal:connect", "terminal:session",
		"linux_user:read", "linux_user:write",
		"project:read", "project:write",
		"ipam:read", "ipam:write",
		"audit:read",
		"approval:read", "approval:review",
		"playbook:read", "playbook:write", "playbook:execute",
		"user:read", "user:write",
		"setting:read", "setting:write",
		"permission:manage",
	},
	"operator": {
		"asset:read", "asset:write", "asset:test_conn", "asset:init",
		"cloud:read", "cloud:write",
		"sshkey:read", "sshkey:write", "sshkey:deploy", "sshkey:revoke",
		"probe:read", "probe:execute",
		"task:read", "task:execute",
		"file:read", "file:distribute",
		"terminal:connect", "terminal:session",
		"linux_user:read", "linux_user:write",
		"project:read", "project:write",
		"ipam:read", "ipam:write",
		"audit:read",
		"approval:read",
		"playbook:read", "playbook:execute",
		"user:read",
		"setting:read",
	},
	"viewer": {
		"asset:read",
		"cloud:read",
		"sshkey:read",
		"probe:read",
		"task:read",
		"file:read",
		"terminal:session",
		"linux_user:read",
		"project:read",
		"ipam:read",
		"audit:read",
		"approval:read",
		"playbook:read",
		"user:read",
		"setting:read",
	},
}

// PermissionService 权限服务
type PermissionService struct {
	repo  *repository.PermissionRepository
	mu    sync.RWMutex
	cache map[string]map[string]bool // role -> permission_code -> true
}

// NewPermissionService 创建 PermissionService
func NewPermissionService(repo *repository.PermissionRepository) *PermissionService {
	return &PermissionService{
		repo:  repo,
		cache: make(map[string]map[string]bool),
	}
}

// SeedDefaults 初始化默认权限和角色映射
func (s *PermissionService) SeedDefaults() error {
	// 种子权限
	for _, p := range defaultPermissions {
		if err := s.repo.CreateIfNotExists(&p); err != nil {
			return err
		}
	}

	// 种子角色权限（仅当角色无权限时才设置默认值）
	for role, codes := range defaultRolePermissions {
		existing, err := s.repo.ListRolePermissions(role)
		if err != nil {
			return err
		}
		if len(existing) == 0 {
			if err := s.repo.SetRolePermissions(role, codes); err != nil {
				return err
			}
		}
	}

	return s.LoadCache()
}

// LoadCache 加载权限缓存
func (s *PermissionService) LoadCache() error {
	rps, err := s.repo.ListAllRolePermissions()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache = make(map[string]map[string]bool)
	for _, rp := range rps {
		if s.cache[rp.Role] == nil {
			s.cache[rp.Role] = make(map[string]bool)
		}
		s.cache[rp.Role][rp.PermissionCode] = true
	}

	logger.Log.Infof("权限缓存已加载: %d 个角色", len(s.cache))
	return nil
}

// HasPermission 检查角色是否有某权限
func (s *PermissionService) HasPermission(role, permCode string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if perms, ok := s.cache[role]; ok {
		return perms[permCode]
	}
	return false
}

// GetRolePermissions 获取角色的所有权限码
func (s *PermissionService) GetRolePermissions(role string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if perms, ok := s.cache[role]; ok {
		codes := make([]string, 0, len(perms))
		for code := range perms {
			codes = append(codes, code)
		}
		return codes
	}
	return nil
}

// ListAllPermissions 列出所有权限定义
func (s *PermissionService) ListAllPermissions() ([]model.Permission, error) {
	return s.repo.ListAll()
}

// ListRolePermissions 列出角色的权限码
func (s *PermissionService) ListRolePermissions(role string) ([]string, error) {
	return s.repo.ListRolePermissions(role)
}

// SetRolePermissions 设置角色权限并刷新缓存
func (s *PermissionService) SetRolePermissions(role string, codes []string) error {
	if err := s.repo.SetRolePermissions(role, codes); err != nil {
		return err
	}
	return s.LoadCache()
}
