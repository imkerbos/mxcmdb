package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/pkg/pagination"
	"github.com/imkerbos/mxcmdb/internal/pkg/response"
	"github.com/imkerbos/mxcmdb/internal/service"
)

// ApprovalHandler 审批处理器
type ApprovalHandler struct {
	svc *service.ApprovalService
}

// NewApprovalHandler 创建 ApprovalHandler
func NewApprovalHandler(svc *service.ApprovalService) *ApprovalHandler {
	return &ApprovalHandler{svc: svc}
}

// Create 创建审批工单
func (h *ApprovalHandler) Create(c *gin.Context) {
	var req dto.CreateApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	userID, _ := c.Get("user_id")
	result, err := h.svc.Create(req, userID.(uint))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, result)
}

// List 审批列表
func (h *ApprovalHandler) List(c *gin.Context) {
	page, pageSize := pagination.Parse(c)
	status := c.Query("status")
	approvalType := c.Query("type")

	items, total, err := h.svc.List(status, approvalType, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询审批列表失败")
		return
	}
	response.Success(c, response.PageData(items, total, page, pageSize))
}

// ListMine 我的审批
func (h *ApprovalHandler) ListMine(c *gin.Context) {
	page, pageSize := pagination.Parse(c)
	userID, _ := c.Get("user_id")

	items, total, err := h.svc.ListMine(userID.(uint), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询审批列表失败")
		return
	}
	response.Success(c, response.PageData(items, total, page, pageSize))
}

// GetByID 审批详情
func (h *ApprovalHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	result, err := h.svc.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, 40401, "审批工单不存在")
		return
	}
	response.Success(c, result)
}

// Review 审批操作
func (h *ApprovalHandler) Review(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	var req dto.ReviewApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	userID, _ := c.Get("user_id")
	result, err := h.svc.Review(uint(id), req, userID.(uint))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, result)
}

// Cancel 取消审批工单
func (h *ApprovalHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	userID, _ := c.Get("user_id")
	if err := h.svc.Cancel(uint(id), userID.(uint)); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, nil)
}

// CountPending 待审批数量
func (h *ApprovalHandler) CountPending(c *gin.Context) {
	count, err := h.svc.CountPending()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询待审批数失败")
		return
	}
	response.Success(c, gin.H{"count": count})
}
