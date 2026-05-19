package service

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/imkerbos/mxcmdb/internal/pkg/logger"
	"github.com/imkerbos/mxcmdb/internal/repository"
)

// HeartbeatService 定期探活服务
type HeartbeatService struct {
	assetRepo *repository.AssetRepository
	configSvc *SystemConfigService
	cron      *cron.Cron
	entryID   cron.EntryID
	mu        sync.Mutex
}

// NewHeartbeatService 创建探活服务
func NewHeartbeatService(assetRepo *repository.AssetRepository, configSvc *SystemConfigService) *HeartbeatService {
	return &HeartbeatService{
		assetRepo: assetRepo,
		configSvc: configSvc,
	}
}

// Start 启动定期探活
func (s *HeartbeatService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	intervalHours := s.configSvc.GetInt("probe.heartbeat_interval_hours", 4)
	spec := fmt.Sprintf("@every %dh", intervalHours)

	s.cron = cron.New()
	var err error
	s.entryID, err = s.cron.AddFunc(spec, s.checkAll)
	if err != nil {
		logger.Log.Errorf("注册探活定时任务失败: %v", err)
		return
	}

	s.cron.Start()
	logger.Log.Infof("探活调度器已启动，间隔 %d 小时", intervalHours)

	// 启动后立即执行一次
	go s.checkAll()
}

// Stop 停止探活调度器
func (s *HeartbeatService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cron != nil {
		s.cron.Stop()
		logger.Log.Info("探活调度器已停止")
	}
}

// checkAll 对所有 IDC 资产执行 TCP 探活
func (s *HeartbeatService) checkAll() {
	assets, err := s.assetRepo.ListBySource("manual")
	if err != nil {
		logger.Log.Errorf("查询 IDC 资产失败: %v", err)
		return
	}
	if len(assets) == 0 {
		return
	}

	timeout := time.Duration(s.configSvc.GetInt("ssh.timeout", 10)) * time.Second
	maxConcurrent := s.configSvc.GetInt("ssh.max_concurrent", 100)

	logger.Log.Infof("开始探活，共 %d 台 IDC 资产", len(assets))

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrent)

	for i := range assets {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			asset := &assets[idx]
			addr := net.JoinHostPort(asset.IP, fmt.Sprintf("%d", asset.Port))
			conn, err := net.DialTimeout("tcp", addr, timeout)

			if err != nil {
				if asset.Status != "offline" {
					asset.Status = "offline"
					_ = s.assetRepo.Update(asset)
				}
			} else {
				conn.Close()
				if asset.Status != "online" {
					asset.Status = "online"
					_ = s.assetRepo.Update(asset)
				}
			}
		}(i)
	}

	wg.Wait()
	logger.Log.Info("探活完成")
}
