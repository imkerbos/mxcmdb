package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/pkg/pagination"
	"github.com/imkerbos/mxcmdb/internal/pkg/response"
	"github.com/imkerbos/mxcmdb/internal/pkg/ws"
	"github.com/imkerbos/mxcmdb/internal/service"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// TaskHandler 批量任务处理器
type TaskHandler struct {
	svc *service.TaskService
	hub *ws.Hub
}

// NewTaskHandler 创建 TaskHandler
func NewTaskHandler(svc *service.TaskService, hub *ws.Hub) *TaskHandler {
	return &TaskHandler{svc: svc, hub: hub}
}

// Execute 创建并执行批量任务
func (h *TaskHandler) Execute(c *gin.Context) {
	var req dto.TaskExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	userID, _ := c.Get("user_id")
	task, err := h.svc.Execute(req, userID.(uint))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, task)
}

// List 任务列表
func (h *TaskHandler) List(c *gin.Context) {
	page, pageSize := pagination.Parse(c)
	status := c.Query("status")

	tasks, total, err := h.svc.List(status, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询任务失败")
		return
	}
	response.Success(c, response.PageData(tasks, total, page, pageSize))
}

// GetByID 获取任务详情
func (h *TaskHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	task, err := h.svc.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, 40400, "任务不存在")
		return
	}
	response.Success(c, task)
}

// GetResults 获取任务执行结果
func (h *TaskHandler) GetResults(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	results, err := h.svc.GetResults(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询任务结果失败")
		return
	}
	response.Success(c, results)
}

// WebSocket 任务实时输出
func (h *TaskHandler) WebSocket(c *gin.Context) {
	id := c.Param("id")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	room := fmt.Sprintf("task_%s", id)
	h.hub.Join(room, conn)
	defer h.hub.Leave(room, conn)

	// 保持连接，等待客户端关闭
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
