package service

import (
	"fmt"

	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// UserService 用户管理服务
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService 创建 UserService
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// List 用户列表
func (s *UserService) List(keyword, role string, status *int, page, pageSize int) ([]dto.UserResponse, int64, error) {
	users, total, err := s.repo.List(keyword, role, status, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %w", err)
	}

	result := make([]dto.UserResponse, len(users))
	for i, u := range users {
		result[i] = toUserResponse(u)
	}
	return result, total, nil
}

// Create 创建用户
func (s *UserService) Create(req dto.CreateUserRequest, createdBy uint) error {
	// 检查用户名是否已存在
	if existing, _ := s.repo.GetByUsername(req.Username); existing != nil {
		return fmt.Errorf("用户名 %s 已存在", req.Username)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	user := &model.User{
		Username:  req.Username,
		Password:  string(hash),
		Nickname:  req.Nickname,
		Email:     req.Email,
		Phone:     req.Phone,
		Role:      req.Role,
		Status:    1,
		CreatedBy: createdBy,
	}
	if err := s.repo.Create(user); err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}
	return nil
}

// Update 更新用户
func (s *UserService) Update(id uint, req dto.UpdateUserRequest) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Role != "" {
		user.Role = req.Role
	}

	return s.repo.Update(user)
}

// Delete 删除用户
func (s *UserService) Delete(id uint) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}
	if user.Username == "admin" {
		return fmt.Errorf("不能删除管理员账号")
	}
	return s.repo.Delete(id)
}

// ResetPassword 重置密码（管理员操作）
func (s *UserService) ResetPassword(id uint, password string) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}
	user.Password = string(hash)
	return s.repo.Update(user)
}

// ChangePassword 修改密码（用户本人操作）
func (s *UserService) ChangePassword(id uint, oldPwd, newPwd string) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPwd)); err != nil {
		return fmt.Errorf("原密码错误")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}
	user.Password = string(hash)
	return s.repo.Update(user)
}

// ToggleStatus 切换用户状态
func (s *UserService) ToggleStatus(id uint, status int) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}
	if user.Username == "admin" {
		return fmt.Errorf("不能禁用管理员账号")
	}
	user.Status = status
	return s.repo.Update(user)
}

// toUserResponse 转换为响应 DTO
func toUserResponse(u model.User) dto.UserResponse {
	resp := dto.UserResponse{
		ID:          u.ID,
		Username:    u.Username,
		Nickname:    u.Nickname,
		Email:       u.Email,
		Phone:       u.Phone,
		Avatar:      u.Avatar,
		Role:        u.Role,
		Status:      u.Status,
		MFAEnabled:  u.MFAEnabled,
		LastLoginIP: u.LastLoginIP,
		CreatedAt:   u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if u.LastLoginAt != nil {
		resp.LastLoginAt = u.LastLoginAt.Format("2006-01-02 15:04:05")
	}
	return resp
}
