package service

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/pkg/logger"
	"github.com/imkerbos/mxcmdb/internal/repository"
)

const auditBufferSize = 1000

// AuditLogService 审计日志服务（异步写入）
type AuditLogService struct {
	repo *repository.AuditLogRepository
	ch   chan model.AuditLog
	done chan struct{}
}

// NewAuditLogService 创建 AuditLogService
func NewAuditLogService(repo *repository.AuditLogRepository) *AuditLogService {
	s := &AuditLogService{
		repo: repo,
		ch:   make(chan model.AuditLog, auditBufferSize),
		done: make(chan struct{}),
	}
	go s.worker()
	return s
}

// Record 异步记录审计日志
func (s *AuditLogService) Record(log model.AuditLog) {
	select {
	case s.ch <- log:
	default:
		logger.Log.Warn("审计日志缓冲区已满，丢弃日志")
	}
}

// List 查询审计日志
func (s *AuditLogService) List(module, action, username, startDate, endDate string, page, pageSize int) ([]model.AuditLog, int64, error) {
	return s.repo.List(module, action, username, startDate, endDate, page, pageSize)
}

// GetByID 根据 ID 查询
func (s *AuditLogService) GetByID(id uint) (*model.AuditLog, error) {
	return s.repo.GetByID(id)
}

// Close 关闭服务，刷新剩余日志
func (s *AuditLogService) Close() {
	close(s.ch)
	<-s.done
}

// worker 后台写入协程
func (s *AuditLogService) worker() {
	defer close(s.done)
	batch := make([]model.AuditLog, 0, 100)

	for log := range s.ch {
		batch = append(batch, log)

		// 批量写入：达到 100 条或 channel 暂时为空
		if len(batch) >= 100 || len(s.ch) == 0 {
			if err := s.repo.BatchCreate(batch); err != nil {
				logger.Log.Errorf("批量写入审计日志失败: %v", err)
			}
			batch = batch[:0]
		}
	}

	// 刷新剩余
	if len(batch) > 0 {
		if err := s.repo.BatchCreate(batch); err != nil {
			logger.Log.Errorf("刷新审计日志失败: %v", err)
		}
	}
}
