package service

import (
	"encoding/json"

	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/repository"
	"gorm.io/gorm"
)

// defaultQuickActions 默认快捷操作
var defaultQuickActions = []string{"probe", "task_execute", "terminal", "cloud_accounts", "assets_idc", "file_distribute"}

// DashboardService 仪表盘服务
type DashboardService struct {
	assetRepo       *repository.AssetRepository
	auditRepo       *repository.AuditLogRepository
	termRepo        *repository.TerminalSessionRepository
	projectRepo     *repository.ProjectRepository
	quickActionRepo *repository.UserQuickActionRepository
}

// NewDashboardService 创建 DashboardService
func NewDashboardService(
	assetRepo *repository.AssetRepository,
	auditRepo *repository.AuditLogRepository,
	termRepo *repository.TerminalSessionRepository,
	projectRepo *repository.ProjectRepository,
	quickActionRepo *repository.UserQuickActionRepository,
) *DashboardService {
	return &DashboardService{
		assetRepo:       assetRepo,
		auditRepo:       auditRepo,
		termRepo:        termRepo,
		projectRepo:     projectRepo,
		quickActionRepo: quickActionRepo,
	}
}

// GetStats 获取仪表盘统计
func (s *DashboardService) GetStats() (*dto.DashboardStats, error) {
	cloudCount, _ := s.assetRepo.CountBySource("cloud_sync")
	idcCount, _ := s.assetRepo.CountBySource("manual")
	onlineCount, _ := s.assetRepo.CountByStatus("online")
	offlineCount, _ := s.assetRepo.CountByStatus("offline")
	probedCount, _ := s.assetRepo.CountProbed()
	projectCount, _ := s.projectRepo.Count()

	stats := &dto.DashboardStats{
		TotalAssets:   cloudCount + idcCount,
		CloudAssets:   cloudCount,
		IDCAssets:     idcCount,
		OnlineAssets:  onlineCount,
		OfflineAssets: offlineCount,
		ProbedAssets:  probedCount,
		TotalProjects: projectCount,
	}

	// 环境分布
	envRows, _ := s.assetRepo.GroupByEnvironment()
	for _, row := range envRows {
		label, _ := row["label"].(string)
		value := toInt64(row["value"])
		if label != "" {
			stats.EnvDistribution = append(stats.EnvDistribution, dto.DistributionItem{Label: label, Value: value})
		}
	}

	// 类型分布
	typeRows, _ := s.assetRepo.GroupByType()
	for _, row := range typeRows {
		label, _ := row["label"].(string)
		value := toInt64(row["value"])
		if label != "" {
			stats.TypeDistribution = append(stats.TypeDistribution, dto.DistributionItem{Label: label, Value: value})
		}
	}

	return stats, nil
}

// GetRecentActivities 获取最近活动
func (s *DashboardService) GetRecentActivities() ([]dto.RecentActivity, error) {
	logs, _, err := s.auditRepo.List("", "", "", "", "", 1, 10)
	if err != nil {
		return nil, err
	}
	activities := make([]dto.RecentActivity, len(logs))
	for i, log := range logs {
		activities[i] = dto.RecentActivity{
			ID:        log.ID,
			Username:  log.Username,
			Module:    log.Module,
			Action:    log.Action,
			Path:      log.Path,
			ClientIP:  log.ClientIP,
			CreatedAt: log.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return activities, nil
}

// GetQuickActions 获取用户快捷操作配置
func (s *DashboardService) GetQuickActions(userID uint) (*dto.QuickActionsResponse, error) {
	record, err := s.quickActionRepo.GetByUserID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &dto.QuickActionsResponse{Actions: defaultQuickActions}, nil
		}
		return nil, err
	}

	var actions []string
	if err := json.Unmarshal([]byte(record.Actions), &actions); err != nil {
		return &dto.QuickActionsResponse{Actions: defaultQuickActions}, nil
	}
	return &dto.QuickActionsResponse{Actions: actions}, nil
}

// UpdateQuickActions 更新用户快捷操作配置
func (s *DashboardService) UpdateQuickActions(userID uint, actions []string) error {
	data, err := json.Marshal(actions)
	if err != nil {
		return err
	}
	return s.quickActionRepo.Upsert(userID, string(data))
}

func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case float64:
		return int64(n)
	case int:
		return int64(n)
	default:
		return 0
	}
}
