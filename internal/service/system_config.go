package service

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/pkg/logger"
	"github.com/imkerbos/mxcmdb/internal/repository"
)

// SystemConfigService 系统配置服务
type SystemConfigService struct {
	repo  *repository.SystemConfigRepository
	cache map[string]string
	mu    sync.RWMutex
}

// NewSystemConfigService 创建 SystemConfigService
func NewSystemConfigService(repo *repository.SystemConfigRepository) *SystemConfigService {
	return &SystemConfigService{
		repo:  repo,
		cache: make(map[string]string),
	}
}

// LoadAll 启动时加载所有配置到内存
func (s *SystemConfigService) LoadAll() error {
	configs, err := s.repo.List()
	if err != nil {
		return fmt.Errorf("load system configs: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, cfg := range configs {
		s.cache[cfg.Key] = cfg.Value
	}
	logger.Log.Infof("已加载 %d 条系统配置", len(configs))
	return nil
}

// Get 获取配置值
func (s *SystemConfigService) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cache[key]
}

// GetWithDefault 获取配置值，不存在时返回默认值
func (s *SystemConfigService) GetWithDefault(key, defaultVal string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.cache[key]; ok {
		return v
	}
	return defaultVal
}

// GetInt 获取整型配置值
func (s *SystemConfigService) GetInt(key string, defaultVal int) int {
	v := s.Get(key)
	if v == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return i
}

// GetBool 获取布尔配置值
func (s *SystemConfigService) GetBool(key string) bool {
	v := s.Get(key)
	return v == "true" || v == "1"
}

// Set 设置配置值并刷新缓存
func (s *SystemConfigService) Set(key, value string) error {
	cfg, err := s.repo.GetByKey(key)
	if err != nil {
		return fmt.Errorf("配置项 %s 不存在", key)
	}
	cfg.Value = value
	if err := s.repo.Upsert(cfg); err != nil {
		return fmt.Errorf("更新配置失败: %w", err)
	}

	s.mu.Lock()
	s.cache[key] = value
	s.mu.Unlock()
	return nil
}

// List 列出所有配置
func (s *SystemConfigService) List() ([]model.SystemConfig, error) {
	return s.repo.List()
}

// ListByCategory 按分类列出配置
func (s *SystemConfigService) ListByCategory(category string) ([]model.SystemConfig, error) {
	return s.repo.GetByCategory(category)
}

// BatchUpdate 批量更新配置
func (s *SystemConfigService) BatchUpdate(items []struct{ Key, Value string }) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range items {
		cfg, err := s.repo.GetByKey(item.Key)
		if err != nil {
			return fmt.Errorf("配置项 %s 不存在", item.Key)
		}
		cfg.Value = item.Value
		if err := s.repo.Upsert(cfg); err != nil {
			return fmt.Errorf("更新配置 %s 失败: %w", item.Key, err)
		}
		s.cache[item.Key] = item.Value
	}
	return nil
}

// SeedDefaults 种子默认配置
func (s *SystemConfigService) SeedDefaults() error {
	defaults := []model.SystemConfig{
		{Key: "ssh.timeout", Value: "30", Category: "ssh", Description: "SSH 连接超时（秒）", Type: "int"},
		{Key: "ssh.max_concurrent", Value: "10", Category: "ssh", Description: "SSH 最大并发数", Type: "int"},
		{Key: "security.dangerous_commands", Value: "rm -rf /,shutdown,reboot,mkfs,dd if=,halt,poweroff", Category: "security", Description: "危险命令黑名单（逗号分隔）", Type: "string"},
		{Key: "security.rate_limit", Value: "100", Category: "security", Description: "接口限流阈值（次/分钟）", Type: "int"},
		{Key: "security.cors_whitelist", Value: "*", Category: "security", Description: "CORS 白名单（逗号分隔）", Type: "string"},
		{Key: "security.mfa_policy", Value: "optional", Category: "security", Description: "MFA 策略（optional=用户自愿 / required_admin=管理员必须 / required_all=全员必须）", Type: "string"},
		{Key: "terminal.idle_timeout", Value: "1800", Category: "terminal", Description: "终端空闲超时（秒）", Type: "int"},
		{Key: "file.max_size", Value: "104857600", Category: "file", Description: "文件上传最大大小（字节，默认100MB）", Type: "int"},
		{Key: "probe.heartbeat_interval_hours", Value: "4", Category: "probe", Description: "IDC 资产探活间隔（小时）", Type: "int"},
	}

	for i := range defaults {
		if err := s.repo.Upsert(&defaults[i]); err != nil {
			return fmt.Errorf("seed config %s: %w", defaults[i].Key, err)
		}
	}
	return nil
}
