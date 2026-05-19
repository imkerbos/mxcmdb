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

// IPAMHandler IPAM 处理器
type IPAMHandler struct {
	svc *service.IPAMService
}

// NewIPAMHandler 创建 IPAMHandler
func NewIPAMHandler(svc *service.IPAMService) *IPAMHandler {
	return &IPAMHandler{svc: svc}
}

// ListSubnets 网段列表
func (h *IPAMHandler) ListSubnets(c *gin.Context) {
	page, pageSize := pagination.Parse(c)

	subnets, total, err := h.svc.ListSubnets(page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询网段失败")
		return
	}
	response.Success(c, response.PageData(subnets, total, page, pageSize))
}

// CreateSubnet 创建网段
func (h *IPAMHandler) CreateSubnet(c *gin.Context) {
	var req dto.CreateSubnetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	subnet, err := h.svc.CreateSubnet(req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, subnet)
}

// UpdateSubnet 更新网段
func (h *IPAMHandler) UpdateSubnet(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	var req dto.UpdateSubnetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	if err := h.svc.UpdateSubnet(uint(id), req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, nil)
}

// DeleteSubnet 删除网段
func (h *IPAMHandler) DeleteSubnet(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	if err := h.svc.DeleteSubnet(uint(id)); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, nil)
}

// ListIPs IP 列表
func (h *IPAMHandler) ListIPs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	page, pageSize := pagination.Parse(c)
	status := c.Query("status")

	ips, total, err := h.svc.ListIPs(uint(id), status, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "查询 IP 失败")
		return
	}
	response.Success(c, response.PageData(ips, total, page, pageSize))
}

// AllocateIP 分配 IP
func (h *IPAMHandler) AllocateIP(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("ipId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	var req dto.AllocateIPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	if err := h.svc.AllocateIP(uint(id), req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, nil)
}

// ReleaseIP 释放 IP
func (h *IPAMHandler) ReleaseIP(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("ipId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	if err := h.svc.ReleaseIP(uint(id)); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.Success(c, nil)
}
