package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/pkg/response"
	"github.com/imkerbos/mxcmdb/internal/service"
)

// ProbeHandler 探针处理器
type ProbeHandler struct {
	svc *service.ProbeService
}

// NewProbeHandler 创建 ProbeHandler
func NewProbeHandler(svc *service.ProbeService) *ProbeHandler {
	return &ProbeHandler{svc: svc}
}

// Execute 触发探针采集
func (h *ProbeHandler) Execute(c *gin.Context) {
	var req dto.ProbeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	if err := h.svc.ProbeAssets(req.AssetIDs); err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, err.Error())
		return
	}
	response.Success(c, nil)
}

// GetLatest 获取最新探针结果
func (h *ProbeHandler) GetLatest(c *gin.Context) {
	assetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	result, err := h.svc.GetLatestResult(uint(assetID))
	if err != nil {
		response.Error(c, http.StatusNotFound, 40400, "暂无探针结果")
		return
	}
	response.Success(c, result)
}

// ListHistory 探针历史
func (h *ProbeHandler) ListHistory(c *gin.Context) {
	assetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	results, err := h.svc.ListResults(uint(assetID), 20)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询探针历史失败")
		return
	}
	response.Success(c, results)
}
