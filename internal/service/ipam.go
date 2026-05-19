package service

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/repository"
)

// IPAMService IP 地址管理服务
type IPAMService struct {
	repo *repository.IPAMRepository
}

// NewIPAMService 创建 IPAMService
func NewIPAMService(repo *repository.IPAMRepository) *IPAMService {
	return &IPAMService{repo: repo}
}

// CreateSubnet 创建网段并初始化 IP 列表
func (s *IPAMService) CreateSubnet(req dto.CreateSubnetRequest) (*dto.SubnetResponse, error) {
	_, ipNet, err := net.ParseCIDR(req.CIDR)
	if err != nil {
		return nil, fmt.Errorf("无效的 CIDR: %w", err)
	}

	// 计算可用 IP 数量（排除网络地址和广播地址）
	ones, bits := ipNet.Mask.Size()
	totalIPs := (1 << (bits - ones)) - 2
	if totalIPs <= 0 {
		return nil, fmt.Errorf("网段太小，无可用 IP")
	}
	if totalIPs > 65534 {
		totalIPs = 65534 // 限制最大 /16
	}

	subnet := &model.Subnet{
		Name:     req.Name,
		CIDR:     req.CIDR,
		Gateway:  req.Gateway,
		VLAN:     req.VLAN,
		TotalIPs: totalIPs,
		UsedIPs:  0,
		Comment:  req.Comment,
	}
	if err := s.repo.CreateSubnet(subnet); err != nil {
		return nil, fmt.Errorf("创建网段失败: %w", err)
	}

	// 初始化 IP 地址列表
	ips := generateIPs(ipNet, subnet.ID)
	if len(ips) > 0 {
		_ = s.repo.BatchCreateIPs(ips)
	}

	return s.toSubnetResponse(subnet), nil
}

// UpdateSubnet 更新网段
func (s *IPAMService) UpdateSubnet(id uint, req dto.UpdateSubnetRequest) error {
	subnet, err := s.repo.GetSubnetByID(id)
	if err != nil {
		return fmt.Errorf("网段不存在")
	}
	if req.Name != "" {
		subnet.Name = req.Name
	}
	if req.Gateway != "" {
		subnet.Gateway = req.Gateway
	}
	if req.VLAN != 0 {
		subnet.VLAN = req.VLAN
	}
	if req.Comment != "" {
		subnet.Comment = req.Comment
	}
	return s.repo.UpdateSubnet(subnet)
}

// DeleteSubnet 删除网段
func (s *IPAMService) DeleteSubnet(id uint) error {
	if _, err := s.repo.GetSubnetByID(id); err != nil {
		return fmt.Errorf("网段不存在")
	}
	_ = s.repo.DeleteIPsBySubnet(id)
	return s.repo.DeleteSubnet(id)
}

// ListSubnets 网段列表
func (s *IPAMService) ListSubnets(page, pageSize int) ([]dto.SubnetResponse, int64, error) {
	subnets, total, err := s.repo.ListSubnets(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.SubnetResponse, len(subnets))
	for i, sub := range subnets {
		used := int(s.repo.CountUsedIPs(sub.ID))
		sub.UsedIPs = used
		result[i] = *s.toSubnetResponse(&sub)
	}
	return result, total, nil
}

// ListIPs 查询 IP 列表
func (s *IPAMService) ListIPs(subnetID uint, status string, page, pageSize int) ([]dto.IPAddressResponse, int64, error) {
	ips, total, err := s.repo.ListIPsBySubnet(subnetID, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.IPAddressResponse, len(ips))
	for i, ip := range ips {
		result[i] = dto.IPAddressResponse{
			ID:       ip.ID,
			SubnetID: ip.SubnetID,
			Address:  ip.Address,
			Status:   ip.Status,
			AssetID:  ip.AssetID,
			Hostname: ip.Hostname,
			Comment:  ip.Comment,
		}
	}
	return result, total, nil
}

// AllocateIP 分配 IP
func (s *IPAMService) AllocateIP(ipID uint, req dto.AllocateIPRequest) error {
	ip, err := s.repo.GetIPByID(ipID)
	if err != nil {
		return fmt.Errorf("IP 不存在")
	}
	if ip.Status == "allocated" {
		return fmt.Errorf("IP 已被分配")
	}
	ip.Status = "allocated"
	ip.AssetID = req.AssetID
	ip.Hostname = req.Hostname
	ip.Comment = req.Comment
	return s.repo.UpdateIP(ip)
}

// ReleaseIP 释放 IP
func (s *IPAMService) ReleaseIP(ipID uint) error {
	ip, err := s.repo.GetIPByID(ipID)
	if err != nil {
		return fmt.Errorf("IP 不存在")
	}
	ip.Status = "available"
	ip.AssetID = nil
	ip.Hostname = ""
	ip.Comment = ""
	return s.repo.UpdateIP(ip)
}

func (s *IPAMService) toSubnetResponse(sub *model.Subnet) *dto.SubnetResponse {
	usage := float64(0)
	if sub.TotalIPs > 0 {
		usage = float64(sub.UsedIPs) / float64(sub.TotalIPs) * 100
	}
	return &dto.SubnetResponse{
		ID:           sub.ID,
		Name:         sub.Name,
		CIDR:         sub.CIDR,
		Gateway:      sub.Gateway,
		VLAN:         sub.VLAN,
		TotalIPs:     sub.TotalIPs,
		UsedIPs:      sub.UsedIPs,
		UsagePercent: usage,
		Comment:      sub.Comment,
	}
}

// generateIPs 根据 CIDR 生成 IP 列表
func generateIPs(ipNet *net.IPNet, subnetID uint) []model.IPAddress {
	var ips []model.IPAddress

	ip := ipNet.IP.To4()
	if ip == nil {
		return ips
	}

	start := binary.BigEndian.Uint32(ip)
	ones, bits := ipNet.Mask.Size()
	total := 1 << (bits - ones)

	// 跳过网络地址(first)和广播地址(last)
	for i := 1; i < total-1; i++ {
		addr := make(net.IP, 4)
		binary.BigEndian.PutUint32(addr, start+uint32(i))
		ips = append(ips, model.IPAddress{
			SubnetID: subnetID,
			Address:  addr.String(),
			Status:   "available",
		})
		if len(ips) >= 65534 {
			break
		}
	}
	return ips
}
