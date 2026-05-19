package repository

import (
	"github.com/imkerbos/mxcmdb/internal/model"
	"gorm.io/gorm"
)

// IPAMRepository IPAM 数据访问
type IPAMRepository struct {
	db *gorm.DB
}

// NewIPAMRepository 创建 IPAMRepository
func NewIPAMRepository(db *gorm.DB) *IPAMRepository {
	return &IPAMRepository{db: db}
}

// CreateSubnet 创建网段
func (r *IPAMRepository) CreateSubnet(subnet *model.Subnet) error {
	return r.db.Create(subnet).Error
}

// UpdateSubnet 更新网段
func (r *IPAMRepository) UpdateSubnet(subnet *model.Subnet) error {
	return r.db.Save(subnet).Error
}

// DeleteSubnet 删除网段
func (r *IPAMRepository) DeleteSubnet(id uint) error {
	return r.db.Delete(&model.Subnet{}, id).Error
}

// GetSubnetByID 按 ID 查询网段
func (r *IPAMRepository) GetSubnetByID(id uint) (*model.Subnet, error) {
	var subnet model.Subnet
	if err := r.db.First(&subnet, id).Error; err != nil {
		return nil, err
	}
	return &subnet, nil
}

// ListSubnets 网段列表
func (r *IPAMRepository) ListSubnets(page, pageSize int) ([]model.Subnet, int64, error) {
	var subnets []model.Subnet
	var total int64

	if err := r.db.Model(&model.Subnet{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.Order("id ASC").Offset(offset).Limit(pageSize).Find(&subnets).Error; err != nil {
		return nil, 0, err
	}
	return subnets, total, nil
}

// CreateIP 创建 IP
func (r *IPAMRepository) CreateIP(ip *model.IPAddress) error {
	return r.db.Create(ip).Error
}

// UpdateIP 更新 IP
func (r *IPAMRepository) UpdateIP(ip *model.IPAddress) error {
	return r.db.Save(ip).Error
}

// GetIPByID 按 ID 查询
func (r *IPAMRepository) GetIPByID(id uint) (*model.IPAddress, error) {
	var ip model.IPAddress
	if err := r.db.First(&ip, id).Error; err != nil {
		return nil, err
	}
	return &ip, nil
}

// ListIPsBySubnet 按网段查询 IP 列表
func (r *IPAMRepository) ListIPsBySubnet(subnetID uint, status string, page, pageSize int) ([]model.IPAddress, int64, error) {
	var ips []model.IPAddress
	var total int64

	query := r.db.Model(&model.IPAddress{}).Where("subnet_id = ?", subnetID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("address ASC").Offset(offset).Limit(pageSize).Find(&ips).Error; err != nil {
		return nil, 0, err
	}
	return ips, total, nil
}

// CountUsedIPs 统计已使用 IP
func (r *IPAMRepository) CountUsedIPs(subnetID uint) int64 {
	var count int64
	r.db.Model(&model.IPAddress{}).Where("subnet_id = ? AND status = ?", subnetID, "allocated").Count(&count)
	return count
}

// BatchCreateIPs 批量创建 IP
func (r *IPAMRepository) BatchCreateIPs(ips []model.IPAddress) error {
	return r.db.CreateInBatches(ips, 100).Error
}

// DeleteIPsBySubnet 删除网段所有 IP
func (r *IPAMRepository) DeleteIPsBySubnet(subnetID uint) error {
	return r.db.Where("subnet_id = ?", subnetID).Delete(&model.IPAddress{}).Error
}
