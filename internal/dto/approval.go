package dto

// CreateApprovalRequest 创建审批工单请求
type CreateApprovalRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required"` // batch_delete / ssh_key_revoke / departure_cleanup / ssh_key_rotate
	Payload     string `json:"payload" binding:"required"`
}

// ReviewApprovalRequest 审批请求
type ReviewApprovalRequest struct {
	Action string `json:"action" binding:"required"` // approve / reject
	Note   string `json:"note"`
}

// ApprovalResponse 审批工单响应
type ApprovalResponse struct {
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Type          string `json:"type"`
	Payload       string `json:"payload"`
	Status        string `json:"status"`
	RequestedBy   uint   `json:"requested_by"`
	RequesterName string `json:"requester_name"`
	ReviewedBy    *uint  `json:"reviewed_by"`
	ReviewerName  string `json:"reviewer_name"`
	ReviewedAt    string `json:"reviewed_at"`
	ReviewNote    string `json:"review_note"`
	CreatedAt     string `json:"created_at"`
}
