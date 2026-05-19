package service

import (
	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/repository"
)

// NotificationService 通知服务
type NotificationService struct {
	repo     *repository.NotificationRepository
	userRepo *repository.UserRepository
}

// NewNotificationService 创建 NotificationService
func NewNotificationService(repo *repository.NotificationRepository, userRepo *repository.UserRepository) *NotificationService {
	return &NotificationService{repo: repo, userRepo: userRepo}
}

// Create 创建通知
func (s *NotificationService) Create(userID uint, title, content, nType, source string, sourceID uint) error {
	n := &model.Notification{
		UserID:   userID,
		Title:    title,
		Content:  content,
		Type:     nType,
		Source:   source,
		SourceID: sourceID,
	}
	return s.repo.Create(n)
}

// Notify 发送通知给指定用户（便捷方法）
func (s *NotificationService) Notify(userID uint, title, content, nType, source string) {
	_ = s.Create(userID, title, content, nType, source, 0)
}

// NotifyAllAdmins 给所有管理员发送通知
func (s *NotificationService) NotifyAllAdmins(title, content, nType, source string) {
	admins, err := s.userRepo.ListByRole("admin")
	if err != nil {
		return
	}
	var notifications []*model.Notification
	for _, admin := range admins {
		notifications = append(notifications, &model.Notification{
			UserID:  admin.ID,
			Title:   title,
			Content: content,
			Type:    nType,
			Source:  source,
		})
	}
	_ = s.repo.CreateBatch(notifications)
}

// NotifyAllOperators 给所有管理员和运维员发送通知
func (s *NotificationService) NotifyAllOperators(title, content, nType, source string) {
	admins, _ := s.userRepo.ListByRole("admin")
	operators, _ := s.userRepo.ListByRole("operator")

	var notifications []*model.Notification
	seen := make(map[uint]bool)

	for _, u := range append(admins, operators...) {
		if seen[u.ID] {
			continue
		}
		seen[u.ID] = true
		notifications = append(notifications, &model.Notification{
			UserID:  u.ID,
			Title:   title,
			Content: content,
			Type:    nType,
			Source:  source,
		})
	}
	_ = s.repo.CreateBatch(notifications)
}

// List 查询通知列表
func (s *NotificationService) List(userID uint, readFilter *bool, page, pageSize int) ([]dto.NotificationResponse, int64, error) {
	items, total, err := s.repo.List(userID, readFilter, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.NotificationResponse, len(items))
	for i, n := range items {
		result[i] = dto.NotificationResponse{
			ID:        n.ID,
			UserID:    n.UserID,
			Title:     n.Title,
			Content:   n.Content,
			Type:      n.Type,
			Source:    n.Source,
			SourceID:  n.SourceID,
			Read:      n.Read,
			CreatedAt: n.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return result, total, nil
}

// CountUnread 查询未读数
func (s *NotificationService) CountUnread(userID uint) (int64, error) {
	return s.repo.CountUnread(userID)
}

// MarkRead 标记已读
func (s *NotificationService) MarkRead(id uint, userID uint) error {
	return s.repo.MarkRead(id, userID)
}

// MarkAllRead 标记全部已读
func (s *NotificationService) MarkAllRead(userID uint) error {
	return s.repo.MarkAllRead(userID)
}

// Delete 删除通知
func (s *NotificationService) Delete(id uint, userID uint) error {
	return s.repo.Delete(id, userID)
}
