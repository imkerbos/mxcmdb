package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/mxcmdb/internal/pkg/pagination"
	"github.com/imkerbos/mxcmdb/internal/pkg/response"
	"github.com/imkerbos/mxcmdb/internal/service"
)

// NotificationHandler 通知处理器
type NotificationHandler struct {
	svc *service.NotificationService
}

// NewNotificationHandler 创建 NotificationHandler
func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// List 通知列表
func (h *NotificationHandler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")
	page, pageSize := pagination.Parse(c)

	var readFilter *bool
	if v := c.Query("read"); v != "" {
		b := v == "true"
		readFilter = &b
	}

	items, total, err := h.svc.List(userID.(uint), readFilter, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询通知失败")
		return
	}
	response.Success(c, response.PageData(items, total, page, pageSize))
}

// UnreadCount 未读数
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID, _ := c.Get("user_id")
	count, err := h.svc.CountUnread(userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询未读数失败")
		return
	}
	response.Success(c, gin.H{"count": count})
}

// MarkRead 标记已读
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}
	userID, _ := c.Get("user_id")
	if err := h.svc.MarkRead(uint(id), userID.(uint)); err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "标记已读失败")
		return
	}
	response.Success(c, nil)
}

// MarkAllRead 标记全部已读
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if err := h.svc.MarkAllRead(userID.(uint)); err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "标记全部已读失败")
		return
	}
	response.Success(c, nil)
}

// Delete 删除通知
func (h *NotificationHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}
	userID, _ := c.Get("user_id")
	if err := h.svc.Delete(uint(id), userID.(uint)); err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "删除通知失败")
		return
	}
	response.Success(c, nil)
}
