package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/pkg/pagination"
	"github.com/imkerbos/mxcmdb/internal/pkg/response"
	"github.com/imkerbos/mxcmdb/internal/service"
)

// FileDistributionHandler 文件分发处理器
type FileDistributionHandler struct {
	svc       *service.FileDistributionService
	configSvc *service.SystemConfigService
}

// NewFileDistributionHandler 创建 FileDistributionHandler
func NewFileDistributionHandler(svc *service.FileDistributionService, configSvc *service.SystemConfigService) *FileDistributionHandler {
	return &FileDistributionHandler{svc: svc, configSvc: configSvc}
}

// Distribute 分发文件
func (h *FileDistributionHandler) Distribute(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "请上传文件")
		return
	}

	// 文件大小校验
	maxSize := int64(h.configSvc.GetInt("file.max_size", 104857600))
	if file.Size > maxSize {
		response.Error(c, http.StatusBadRequest, 40001, "文件大小超出限制")
		return
	}

	// 解析分发参数
	reqJSON := c.PostForm("data")
	var req dto.FileDistributeRequest
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	// 保存临时文件
	tmpDir := os.TempDir()
	tmpPath := filepath.Join(tmpDir, file.Filename)
	if err := c.SaveUploadedFile(file, tmpPath); err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "保存文件失败")
		return
	}

	userID, _ := c.Get("user_id")
	task, err := h.svc.Distribute(tmpPath, file.Filename, file.Size, req, userID.(uint))
	if err != nil {
		os.Remove(tmpPath)
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, task)
}

// List 文件任务列表
func (h *FileDistributionHandler) List(c *gin.Context) {
	page, pageSize := pagination.Parse(c)

	tasks, total, err := h.svc.List(page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询文件任务失败")
		return
	}
	response.Success(c, response.PageData(tasks, total, page, pageSize))
}

// GetResults 获取分发结果
func (h *FileDistributionHandler) GetResults(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	results, err := h.svc.GetResults(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询结果失败")
		return
	}
	response.Success(c, results)
}
