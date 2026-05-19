package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/mxcmdb/internal/pkg/pagination"
	"github.com/imkerbos/mxcmdb/internal/pkg/response"
	"github.com/imkerbos/mxcmdb/internal/service"
)

// TerminalHandler Web Terminal 处理器
type TerminalHandler struct {
	svc *service.TerminalService
}

// NewTerminalHandler 创建 TerminalHandler
func NewTerminalHandler(svc *service.TerminalService) *TerminalHandler {
	return &TerminalHandler{svc: svc}
}

// WebSocket 终端 WebSocket 连接
func (h *TerminalHandler) WebSocket(c *gin.Context) {
	assetID, err := strconv.ParseUint(c.Param("assetId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	h.svc.HandleWebSocket(conn, uint(assetID), userID.(uint), username.(string), c.ClientIP())
}

// ListSessions 会话列表（支持 status 过滤）
func (h *TerminalHandler) ListSessions(c *gin.Context) {
	page, pageSize := pagination.Parse(c)
	status := c.Query("status")

	sessions, total, err := h.svc.ListSessions(page, pageSize, status)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询会话失败")
		return
	}
	response.Success(c, response.PageData(sessions, total, page, pageSize))
}

// GetRecording 获取会话录制
func (h *TerminalHandler) GetRecording(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	recording, err := h.svc.GetRecording(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, 40400, "会话不存在")
		return
	}

	response.Success(c, recording)
}

// ListActiveSessions 在线会话列表
func (h *TerminalHandler) ListActiveSessions(c *gin.Context) {
	result := h.svc.ListActiveSessions()
	response.Success(c, result)
}

// KillSession 强制终止会话
func (h *TerminalHandler) KillSession(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	if err := h.svc.KillSession(uint(id)); err != nil {
		response.Error(c, http.StatusNotFound, 40400, err.Error())
		return
	}
	response.Success(c, nil)
}

