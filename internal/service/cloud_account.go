package service

import (
	"fmt"
	"strings"

	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/pkg/crypto"
	"github.com/imkerbos/mxcmdb/internal/repository"
)

// CloudAccountService 云账号服务
type CloudAccountService struct {
	repo      *repository.CloudAccountRepository
	masterKey string
}

// NewCloudAccountService 创建 CloudAccountService
func NewCloudAccountService(repo *repository.CloudAccountRepository, masterKey string) *CloudAccountService {
	return &CloudAccountService{repo: repo, masterKey: masterKey}
}

// List 云账号列表
func (s *CloudAccountService) List(keyword, provider string, status *int, page, pageSize int) ([]dto.CloudAccountResponse, int64, error) {
	accounts, total, err := s.repo.List(keyword, provider, status, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("查询云账号列表失败: %w", err)
	}

	result := make([]dto.CloudAccountResponse, len(accounts))
	for i, a := range accounts {
		result[i] = s.toResponse(a)
	}
	return result, total, nil
}

// GetByID 获取云账号详情
func (s *CloudAccountService) GetByID(id uint) (*dto.CloudAccountResponse, error) {
	account, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("云账号不存在")
	}
	resp := s.toResponse(*account)
	return &resp, nil
}

// Create 创建云账号
func (s *CloudAccountService) Create(req dto.CreateCloudAccountRequest, createdBy uint) error {
	encrypted, err := crypto.Encrypt(req.AccessKeySecret, s.masterKey)
	if err != nil {
		return fmt.Errorf("加密 AccessKeySecret 失败: %w", err)
	}

	account := &model.CloudAccount{
		Name:            req.Name,
		Provider:        req.Provider,
		AccessKeyID:     req.AccessKeyID,
		AccessKeySecret: encrypted,
		Region:          req.Region,
		Status:          1,
		CreatedBy:       createdBy,
	}
	return s.repo.Create(account)
}

// Update 更新云账号
func (s *CloudAccountService) Update(id uint, req dto.UpdateCloudAccountRequest) error {
	account, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("云账号不存在")
	}

	if req.Name != "" {
		account.Name = req.Name
	}
	if req.AccessKeyID != "" {
		account.AccessKeyID = req.AccessKeyID
	}
	if req.AccessKeySecret != "" {
		encrypted, err := crypto.Encrypt(req.AccessKeySecret, s.masterKey)
		if err != nil {
			return fmt.Errorf("加密 AccessKeySecret 失败: %w", err)
		}
		account.AccessKeySecret = encrypted
	}
	if req.Region != "" {
		account.Region = req.Region
	}

	return s.repo.Update(account)
}

// Delete 删除云账号
func (s *CloudAccountService) Delete(id uint) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return fmt.Errorf("云账号不存在")
	}
	return s.repo.Delete(id)
}

// ToggleStatus 切换状态
func (s *CloudAccountService) ToggleStatus(id uint, status int) error {
	account, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("云账号不存在")
	}
	account.Status = status
	return s.repo.Update(account)
}

// GetDecryptedSecret 获取解密后的 AccessKeySecret（内部使用）
func (s *CloudAccountService) GetDecryptedSecret(id uint) (string, error) {
	account, err := s.repo.GetByID(id)
	if err != nil {
		return "", fmt.Errorf("云账号不存在")
	}
	return crypto.Decrypt(account.AccessKeySecret, s.masterKey)
}

// toResponse 转换为响应 DTO（脱敏）
func (s *CloudAccountService) toResponse(a model.CloudAccount) dto.CloudAccountResponse {
	resp := dto.CloudAccountResponse{
		ID:              a.ID,
		Name:            a.Name,
		Provider:        a.Provider,
		AccessKeyID:     a.AccessKeyID,
		AccessKeySecret: maskSecret(a.AccessKeyID),
		Region:          a.Region,
		Status:          a.Status,
		LastSyncStatus:  a.LastSyncStatus,
		CreatedAt:       a.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if a.LastSyncAt != nil {
		resp.LastSyncAt = a.LastSyncAt.Format("2006-01-02 15:04:05")
	}
	return resp
}

// maskSecret 脱敏显示
func maskSecret(s string) string {
	if len(s) <= 6 {
		return "******"
	}
	return s[:3] + strings.Repeat("*", len(s)-6) + s[len(s)-3:]
}
