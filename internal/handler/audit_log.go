package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/mxcmdb/internal/pkg/pagination"
	"github.com/imkerbos/mxcmdb/internal/pkg/response"
	"github.com/imkerbos/mxcmdb/internal/service"
)

// AuditLogHandler 审计日志处理器
type AuditLogHandler struct {
	svc *service.AuditLogService
}

// NewAuditLogHandler 创建 AuditLogHandler
func NewAuditLogHandler(svc *service.AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{svc: svc}
}

// List 查询审计日志列表
func (h *AuditLogHandler) List(c *gin.Context) {
	page, pageSize := pagination.Parse(c)
	module := c.Query("module")
	action := c.Query("action")
	username := c.Query("username")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	logs, total, err := h.svc.List(module, action, username, startDate, endDate, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询审计日志失败")
		return
	}

	response.Success(c, response.PageData(logs, total, page, pageSize))
}

// GetByID 查询审计日志详情
func (h *AuditLogHandler) GetByID(c *gin.Context) {
	var uri struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	log, err := h.svc.GetByID(uri.ID)
	if err != nil {
		response.Error(c, http.StatusNotFound, 40400, "审计日志不存在")
		return
	}
	response.Success(c, log)
}
