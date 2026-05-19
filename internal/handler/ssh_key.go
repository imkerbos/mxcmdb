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

// SSHKeyHandler SSH Key 处理器
type SSHKeyHandler struct {
	svc *service.SSHKeyService
}

// NewSSHKeyHandler 创建 SSHKeyHandler
func NewSSHKeyHandler(svc *service.SSHKeyService) *SSHKeyHandler {
	return &SSHKeyHandler{svc: svc}
}

// List SSH Key 列表
func (h *SSHKeyHandler) List(c *gin.Context) {
	page, pageSize := pagination.Parse(c)
	keyword := c.Query("keyword")

	keys, total, err := h.svc.List(keyword, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询 SSH Key 失败")
		return
	}
	response.Success(c, response.PageData(keys, total, page, pageSize))
}

// Create 创建 SSH Key
func (h *SSHKeyHandler) Create(c *gin.Context) {
	var req dto.CreateSSHKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	createdBy, _ := c.Get("user_id")
	key, err := h.svc.Create(req, createdBy.(uint))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, key)
}

// Delete 删除 SSH Key
func (h *SSHKeyHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}
	if err := h.svc.Delete(uint(id)); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, nil)
}

// Deploy 部署 SSH Key
func (h *SSHKeyHandler) Deploy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	var req dto.DeploySSHKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	operatorID, _ := c.Get("user_id")
	if err := h.svc.Deploy(uint(id), req, operatorID.(uint)); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, nil)
}

// Download 下载 SSH Key（含解密私钥）
func (h *SSHKeyHandler) Download(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	result, err := h.svc.Download(uint(id))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, result)
}

// ListBindings 查询绑定关系
func (h *SSHKeyHandler) ListBindings(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	bindings, err := h.svc.ListBindings(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询绑定失败")
		return
	}
	response.Success(c, bindings)
}

// Revoke 撤销（回收）SSH Key
func (h *SSHKeyHandler) Revoke(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	var req dto.RevokeSSHKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	operatorID, _ := c.Get("user_id")
	result, err := h.svc.Revoke(uint(id), req, operatorID.(uint))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, result)
}

// Rotate SSH Key 轮换
func (h *SSHKeyHandler) Rotate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	var req dto.RotateSSHKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	createdBy, _ := c.Get("user_id")
	result, err := h.svc.Rotate(uint(id), req, createdBy.(uint))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, result)
}

// DepartureCleanup 离职清理
func (h *SSHKeyHandler) DepartureCleanup(c *gin.Context) {
	var req dto.DepartureCleanupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	operatorID, _ := c.Get("user_id")
	result, err := h.svc.DepartureCleanup(req, operatorID.(uint))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, result)
}

// ListDeployLogs 查询部署日志
func (h *SSHKeyHandler) ListDeployLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	page, pageSize := pagination.Parse(c)
	logs, total, err := h.svc.ListDeployLogs(uint(id), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询部署日志失败")
		return
	}
	response.Success(c, response.PageData(logs, total, page, pageSize))
}
