package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/mxcmdb/internal/pkg/response"
	"github.com/imkerbos/mxcmdb/internal/service"
)

// PermissionHandler 权限管理处理器
type PermissionHandler struct {
	svc *service.PermissionService
}

// NewPermissionHandler 创建 PermissionHandler
func NewPermissionHandler(svc *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{svc: svc}
}

// ListPermissions 列出所有权限定义
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	perms, err := h.svc.ListAllPermissions()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询权限失败")
		return
	}
	response.Success(c, perms)
}

// GetRolePermissions 查询角色权限
func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	role := c.Param("role")
	codes, err := h.svc.ListRolePermissions(role)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询角色权限失败")
		return
	}
	response.Success(c, gin.H{"role": role, "permissions": codes})
}

// SetRolePermissions 设置角色权限
func (h *PermissionHandler) SetRolePermissions(c *gin.Context) {
	role := c.Param("role")

	var req struct {
		Permissions []string `json:"permissions" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	if err := h.svc.SetRolePermissions(role, req.Permissions); err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "设置角色权限失败")
		return
	}
	response.Success(c, nil)
}

// GetMyPermissions 获取当前用户的权限列表
func (h *PermissionHandler) GetMyPermissions(c *gin.Context) {
	role, _ := c.Get("role")
	codes := h.svc.GetRolePermissions(role.(string))
	response.Success(c, gin.H{"role": role, "permissions": codes})
}
