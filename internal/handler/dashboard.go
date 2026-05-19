package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/pkg/response"
	"github.com/imkerbos/mxcmdb/internal/service"
)

// DashboardHandler 仪表盘处理器
type DashboardHandler struct {
	svc *service.DashboardService
}

// NewDashboardHandler 创建 DashboardHandler
func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// GetStats 获取统计数据
func (h *DashboardHandler) GetStats(c *gin.Context) {
	stats, err := h.svc.GetStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "获取统计数据失败")
		return
	}
	response.Success(c, stats)
}

// GetRecentActivities 获取最近活动
func (h *DashboardHandler) GetRecentActivities(c *gin.Context) {
	activities, err := h.svc.GetRecentActivities()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "获取活动失败")
		return
	}
	response.Success(c, activities)
}

// GetQuickActions 获取用户快捷操作配置
func (h *DashboardHandler) GetQuickActions(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	result, err := h.svc.GetQuickActions(userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "获取快捷操作失败")
		return
	}
	response.Success(c, result)
}

// UpdateQuickActions 更新用户快捷操作配置
func (h *DashboardHandler) UpdateQuickActions(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	var req dto.UpdateQuickActionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}
	if len(req.Actions) > 12 {
		response.Error(c, http.StatusBadRequest, 40001, "快捷操作数量不能超过 12 个")
		return
	}
	if err := h.svc.UpdateQuickActions(userID, req.Actions); err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "更新快捷操作失败")
		return
	}
	response.Success(c, nil)
}
