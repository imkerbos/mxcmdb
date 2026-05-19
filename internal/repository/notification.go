package repository

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
)

// NotificationRepository 通知数据访问层
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 创建 NotificationRepository
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create 创建通知
func (r *NotificationRepository) Create(n *model.Notification) error {
	return r.db.Create(n).Error
}

// List 分页查询用户的通知
func (r *NotificationRepository) List(userID uint, readFilter *bool, page, pageSize int) ([]model.Notification, int64, error) {
	var items []model.Notification
	var total int64

	query := r.db.Model(&model.Notification{}).Where("user_id = ?", userID)
	if readFilter != nil {
		query = query.Where("read = ?", *readFilter)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// CountUnread 查询未读数
func (r *NotificationRepository) CountUnread(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).Where("user_id = ? AND read = ?", userID, false).Count(&count).Error
	return count, err
}

// MarkRead 标记已读
func (r *NotificationRepository) MarkRead(id uint, userID uint) error {
	return r.db.Model(&model.Notification{}).Where("id = ? AND user_id = ?", id, userID).Update("read", true).Error
}

// MarkAllRead 标记全部已读
func (r *NotificationRepository) MarkAllRead(userID uint) error {
	return r.db.Model(&model.Notification{}).Where("user_id = ? AND read = ?", userID, false).Update("read", true).Error
}

// Delete 删除通知
func (r *NotificationRepository) Delete(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Notification{}).Error
}

// CreateBatch 批量创建通知（给多个用户发送相同通知）
func (r *NotificationRepository) CreateBatch(notifications []*model.Notification) error {
	if len(notifications) == 0 {
		return nil
	}
	return r.db.Create(&notifications).Error
}
