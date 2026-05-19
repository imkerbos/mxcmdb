package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/pkg/response"
	"github.com/imkerbos/mxcmdb/internal/pkg/ws"
	"github.com/imkerbos/mxcmdb/internal/service"
)

// PlaybookHandler 剧本处理器
type PlaybookHandler struct {
	svc *service.PlaybookService
	hub *ws.Hub
}

// NewPlaybookHandler 创建 PlaybookHandler
func NewPlaybookHandler(svc *service.PlaybookService, hub *ws.Hub) *PlaybookHandler {
	return &PlaybookHandler{svc: svc, hub: hub}
}

// List 查询剧本列表
func (h *PlaybookHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询剧本失败")
		return
	}
	response.Success(c, list)
}

// GetByID 查询剧本详情
func (h *PlaybookHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	pb, err := h.svc.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, 40401, err.Error())
		return
	}
	response.Success(c, pb)
}

// Create 创建自定义剧本
func (h *PlaybookHandler) Create(c *gin.Context) {
	var req dto.CreatePlaybookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	userID, _ := c.Get("user_id")
	pb, err := h.svc.Create(req, userID.(uint))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, pb)
}

// Update 更新剧本
func (h *PlaybookHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	var req dto.UpdatePlaybookRequest
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

// Delete 删除剧本
func (h *PlaybookHandler) Delete(c *gin.Context) {
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

// Execute 执行剧本
func (h *PlaybookHandler) Execute(c *gin.Context) {
	var req dto.ExecutePlaybookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	userID, _ := c.Get("user_id")
	result, err := h.svc.Execute(req, userID.(uint))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, result)
}

// ListExecutions 查询执行历史
func (h *PlaybookHandler) ListExecutions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := h.svc.ListExecutions(page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询执行历史失败")
		return
	}
	response.Success(c, response.PageData(items, total, page, pageSize))
}

// GetExecutionResults 查询执行结果
func (h *PlaybookHandler) GetExecutionResults(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	results, err := h.svc.GetExecutionResults(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询执行结果失败")
		return
	}
	response.Success(c, results)
}

// WebSocket 剧本执行实时输出
func (h *PlaybookHandler) WebSocket(c *gin.Context) {
	id := c.Param("id")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	h.hub.Join(id, conn)
	defer h.hub.Leave(id, conn)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
