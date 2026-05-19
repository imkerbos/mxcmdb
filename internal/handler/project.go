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

// ProjectHandler 项目管理处理器
type ProjectHandler struct {
	svc *service.ProjectService
}

// NewProjectHandler 创建 ProjectHandler
func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

// List 项目列表
func (h *ProjectHandler) List(c *gin.Context) {
	page, pageSize := pagination.Parse(c)
	keyword := c.Query("keyword")
	status := c.Query("status")

	projects, total, err := h.svc.List(keyword, status, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询项目失败")
		return
	}
	response.Success(c, response.PageData(projects, total, page, pageSize))
}

// GetByID 项目详情
func (h *ProjectHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	project, err := h.svc.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, 40400, err.Error())
		return
	}
	response.Success(c, project)
}

// Create 创建项目
func (h *ProjectHandler) Create(c *gin.Context) {
	var req dto.CreateProjectRequest
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

// Update 更新项目
func (h *ProjectHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	var req dto.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	if err := h.svc.Update(uint(id), req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, nil)
}

// Delete 删除项目
func (h *ProjectHandler) Delete(c *gin.Context) {
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

// ListAll 获取所有项目（用于选择器）
func (h *ProjectHandler) ListAll(c *gin.Context) {
	projects, err := h.svc.ListAll()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询项目失败")
		return
	}
	response.Success(c, projects)
}

// GetSummary 项目资产指纹聚合
func (h *ProjectHandler) GetSummary(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	summary, err := h.svc.GetSummary(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, 40400, err.Error())
		return
	}
	response.Success(c, summary)
}

// ListAssets 项目下资产列表
func (h *ProjectHandler) ListAssets(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	page, pageSize := pagination.Parse(c)
	assets, total, err := h.svc.ListAssets(uint(id), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusNotFound, 40400, err.Error())
		return
	}
	response.Success(c, response.PageData(assets, total, page, pageSize))
}
