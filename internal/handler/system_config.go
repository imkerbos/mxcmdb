package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/pkg/response"
	"github.com/imkerbos/mxcmdb/internal/service"
)

// SystemConfigHandler 系统配置处理器
type SystemConfigHandler struct {
	svc *service.SystemConfigService
}

// NewSystemConfigHandler 创建 SystemConfigHandler
func NewSystemConfigHandler(svc *service.SystemConfigService) *SystemConfigHandler {
	return &SystemConfigHandler{svc: svc}
}

// List 获取配置列表
func (h *SystemConfigHandler) List(c *gin.Context) {
	category := c.Query("category")

	var configs interface{}
	var err error
	if category != "" {
		configs, err = h.svc.ListByCategory(category)
	} else {
		configs, err = h.svc.List()
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "获取配置失败")
		return
	}
	response.Success(c, configs)
}

// Update 更新单条配置
func (h *SystemConfigHandler) Update(c *gin.Context) {
	key := c.Param("key")
	var req dto.SystemConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	if err := h.svc.Set(key, req.Value); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, nil)
}

// BatchUpdate 批量更新配置
func (h *SystemConfigHandler) BatchUpdate(c *gin.Context) {
	var req dto.SystemConfigBatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	items := make([]struct{ Key, Value string }, len(req.Configs))
	for i, cfg := range req.Configs {
		items[i] = struct{ Key, Value string }{Key: cfg.Key, Value: cfg.Value}
	}

	if err := h.svc.BatchUpdate(items); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, nil)
}
