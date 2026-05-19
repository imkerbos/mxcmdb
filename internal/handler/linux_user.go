package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/pkg/pagination"
	"github.com/imkerbos/mxcmdb/internal/pkg/response"
	"github.com/imkerbos/mxcmdb/internal/service"
)

// LinuxUserHandler Linux 用户处理器
type LinuxUserHandler struct {
	svc *service.LinuxUserService
}

// NewLinuxUserHandler 创建 LinuxUserHandler
func NewLinuxUserHandler(svc *service.LinuxUserService) *LinuxUserHandler {
	return &LinuxUserHandler{svc: svc}
}

// List Linux 用户列表
func (h *LinuxUserHandler) List(c *gin.Context) {
	page, pageSize := pagination.Parse(c)
	username := c.Query("username")

	users, total, err := h.svc.List(username, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询 Linux 用户失败")
		return
	}
	response.Success(c, response.PageData(users, total, page, pageSize))
}

// Create 创建 Linux 用户
func (h *LinuxUserHandler) Create(c *gin.Context) {
	var req dto.CreateLinuxUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	if err := h.svc.Create(req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, nil)
}

// Delete 删除 Linux 用户
func (h *LinuxUserHandler) Delete(c *gin.Context) {
	username := c.Param("username")
	var req dto.DeleteLinuxUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	if err := h.svc.Delete(username, req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, nil)
}
