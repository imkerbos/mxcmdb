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

// CloudAccountHandler 云账号处理器
type CloudAccountHandler struct {
	svc *service.CloudAccountService
}

// NewCloudAccountHandler 创建 CloudAccountHandler
func NewCloudAccountHandler(svc *service.CloudAccountService) *CloudAccountHandler {
	return &CloudAccountHandler{svc: svc}
}

// List 云账号列表
func (h *CloudAccountHandler) List(c *gin.Context) {
	page, pageSize := pagination.Parse(c)
	keyword := c.Query("keyword")
	provider := c.Query("provider")
	var status *int
	if s := c.Query("status"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			status = &v
		}
	}

	accounts, total, err := h.svc.List(keyword, provider, status, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询云账号失败")
		return
	}

	response.Success(c, response.PageData(accounts, total, page, pageSize))
}

// GetByID 云账号详情
func (h *CloudAccountHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	account, err := h.svc.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, 40400, err.Error())
		return
	}
	response.Success(c, account)
}

// Create 创建云账号
func (h *CloudAccountHandler) Create(c *gin.Context) {
	var req dto.CreateCloudAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	createdBy, _ := c.Get("user_id")
	if err := h.svc.Create(req, createdBy.(uint)); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, nil)
}

// Update 更新云账号
func (h *CloudAccountHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	var req dto.UpdateCloudAccountRequest
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

// Delete 删除云账号
func (h *CloudAccountHandler) Delete(c *gin.Context) {
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
