package service

import (
	"fmt"
	"time"

	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/pkg/logger"
	"github.com/imkerbos/mxcmdb/internal/repository"
)

// ApprovalService 审批服务
type ApprovalService struct {
	repo      *repository.ApprovalRepository
	userRepo  *repository.UserRepository
	notifySvc *NotificationService
}

// NewApprovalService 创建 ApprovalService
func NewApprovalService(repo *repository.ApprovalRepository, userRepo *repository.UserRepository, notifySvc *NotificationService) *ApprovalService {
	return &ApprovalService{
		repo:      repo,
		userRepo:  userRepo,
		notifySvc: notifySvc,
	}
}

// Create 创建审批工单
func (s *ApprovalService) Create(req dto.CreateApprovalRequest, requestedBy uint) (*dto.ApprovalResponse, error) {
	user, err := s.userRepo.GetByID(requestedBy)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	approval := &model.Approval{
		Title:         req.Title,
		Description:   req.Description,
		Type:          req.Type,
		Payload:       req.Payload,
		Status:        "pending",
		RequestedBy:   requestedBy,
		RequesterName: user.Nickname,
	}

	if err := s.repo.Create(approval); err != nil {
		return nil, fmt.Errorf("创建审批工单失败: %w", err)
	}

	// 通知管理员
	if s.notifySvc != nil {
		s.notifySvc.NotifyAllAdmins(
			"新的审批请求",
			fmt.Sprintf("%s 提交了审批请求：%s", user.Nickname, req.Title),
			"warning",
			"system",
		)
	}

	return s.toResponse(approval), nil
}

// Review 审批工单
func (s *ApprovalService) Review(id uint, req dto.ReviewApprovalRequest, reviewedBy uint) (*dto.ApprovalResponse, error) {
	approval, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("审批工单不存在")
	}

	if approval.Status != "pending" {
		return nil, fmt.Errorf("该工单已%s，不可重复审批", approval.Status)
	}

	reviewer, err := s.userRepo.GetByID(reviewedBy)
	if err != nil {
		return nil, fmt.Errorf("审批人不存在")
	}

	now := time.Now()
	approval.ReviewedBy = &reviewedBy
	approval.ReviewerName = reviewer.Nickname
	approval.ReviewedAt = &now
	approval.ReviewNote = req.Note

	switch req.Action {
	case "approve":
		approval.Status = "approved"
	case "reject":
		approval.Status = "rejected"
	default:
		return nil, fmt.Errorf("无效的审批操作: %s", req.Action)
	}

	if err := s.repo.Update(approval); err != nil {
		return nil, fmt.Errorf("更新审批工单失败: %w", err)
	}

	// 通知发起人
	if s.notifySvc != nil {
		nType := "success"
		statusText := "通过"
		if approval.Status == "rejected" {
			nType = "error"
			statusText = "驳回"
		}
		s.notifySvc.Notify(
			approval.RequestedBy,
			fmt.Sprintf("审批%s：%s", statusText, approval.Title),
			fmt.Sprintf("%s 已%s您的审批请求。%s", reviewer.Nickname, statusText, req.Note),
			nType,
			"system",
		)
	}

	logger.Log.Infof("审批工单 %d %s by %s", id, req.Action, reviewer.Username)
	return s.toResponse(approval), nil
}

// Cancel 取消审批工单（发起人操作）
func (s *ApprovalService) Cancel(id uint, userID uint) error {
	approval, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("审批工单不存在")
	}
	if approval.RequestedBy != userID {
		return fmt.Errorf("只能取消自己的审批工单")
	}
	if approval.Status != "pending" {
		return fmt.Errorf("只能取消待审批的工单")
	}

	approval.Status = "cancelled"
	return s.repo.Update(approval)
}

// List 查询审批列表（管理员查看全部，普通用户查看自己的）
func (s *ApprovalService) List(status, approvalType string, page, pageSize int) ([]dto.ApprovalResponse, int64, error) {
	items, total, err := s.repo.List(status, approvalType, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.ApprovalResponse, len(items))
	for i, a := range items {
		result[i] = *s.toResponse(&a)
	}
	return result, total, nil
}

// ListMine 查询我发起的审批
func (s *ApprovalService) ListMine(userID uint, page, pageSize int) ([]dto.ApprovalResponse, int64, error) {
	items, total, err := s.repo.ListByRequester(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.ApprovalResponse, len(items))
	for i, a := range items {
		result[i] = *s.toResponse(&a)
	}
	return result, total, nil
}

// CountPending 统计待审批数量
func (s *ApprovalService) CountPending() (int64, error) {
	return s.repo.CountPending()
}

// GetByID 查询工单详情
func (s *ApprovalService) GetByID(id uint) (*dto.ApprovalResponse, error) {
	a, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.toResponse(a), nil
}

func (s *ApprovalService) toResponse(a *model.Approval) *dto.ApprovalResponse {
	resp := &dto.ApprovalResponse{
		ID:            a.ID,
		Title:         a.Title,
		Description:   a.Description,
		Type:          a.Type,
		Payload:       a.Payload,
		Status:        a.Status,
		RequestedBy:   a.RequestedBy,
		RequesterName: a.RequesterName,
		ReviewedBy:    a.ReviewedBy,
		ReviewerName:  a.ReviewerName,
		ReviewNote:    a.ReviewNote,
		CreatedAt:     a.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if a.ReviewedAt != nil {
		resp.ReviewedAt = a.ReviewedAt.Format("2006-01-02 15:04:05")
	}
	return resp
}
