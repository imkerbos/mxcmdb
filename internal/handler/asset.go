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

// AssetHandler 资产管理处理器
type AssetHandler struct {
	svc *service.AssetService
}

// NewAssetHandler 创建 AssetHandler
func NewAssetHandler(svc *service.AssetService) *AssetHandler {
	return &AssetHandler{svc: svc}
}

// List 资产列表
func (h *AssetHandler) List(c *gin.Context) {
	page, pageSize := pagination.Parse(c)

	var projectID uint
	if pid := c.Query("project_id"); pid != "" {
		if v, err := strconv.ParseUint(pid, 10, 64); err == nil {
			projectID = uint(v)
		}
	}

	assets, total, err := h.svc.List(
		c.Query("keyword"),
		c.Query("type"),
		c.Query("source"),
		c.Query("status"),
		c.Query("environment"),
		c.Query("department"),
		projectID,
		c.Query("owner"),
		page, pageSize,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询资产列表失败")
		return
	}

	response.Success(c, response.PageData(assets, total, page, pageSize))
}

// GetByID 资产详情
func (h *AssetHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	asset, err := h.svc.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, 40400, err.Error())
		return
	}
	response.Success(c, asset)
}

// GetDetail 资产详情（聚合探针、SSH Key、Linux 用户、终端会话）
func (h *AssetHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	detail, err := h.svc.GetDetail(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, 40400, err.Error())
		return
	}
	response.Success(c, detail)
}

// Create 创建资产
func (h *AssetHandler) Create(c *gin.Context) {
	var req dto.CreateAssetRequest
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

// Update 更新资产
func (h *AssetHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	var req dto.UpdateAssetRequest
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

// Delete 删除资产
func (h *AssetHandler) Delete(c *gin.Context) {
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

// Stats 资产统计
func (h *AssetHandler) Stats(c *gin.Context) {
	stats, err := h.svc.GetStats()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "获取统计失败")
		return
	}
	response.Success(c, stats)
}

// ListAll 获取所有资产（用于选择器）
func (h *AssetHandler) ListAll(c *gin.Context) {
	assets, err := h.svc.ListAll()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "获取资产失败")
		return
	}
	response.Success(c, assets)
}

// BatchImport 批量导入资产
func (h *AssetHandler) BatchImport(c *gin.Context) {
	var req dto.BatchImportAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误：请提供 assets 数组")
		return
	}

	result, err := h.svc.BatchImport(req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "批量导入失败")
		return
	}
	response.Success(c, result)
}

// TestConnection 批量测试 SSH 连接
func (h *AssetHandler) TestConnection(c *gin.Context) {
	var req dto.BatchTestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误：请提供 asset_ids 数组")
		return
	}

	result, err := h.svc.TestConnection(req.AssetIDs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "测试连接失败")
		return
	}
	response.Success(c, result)
}

// BatchDelete 批量删除资产
func (h *AssetHandler) BatchDelete(c *gin.Context) {
	var req dto.BatchDeleteAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误：请提供 ids 数组")
		return
	}

	result, err := h.svc.BatchDelete(req.IDs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "批量删除失败")
		return
	}
	response.Success(c, result)
}
