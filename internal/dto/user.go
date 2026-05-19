package dto

// UserListRequest 用户列表请求
type UserListRequest struct {
	Keyword  string `form:"keyword"`
	Role     string `form:"role"`
	Status   *int   `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Nickname string `json:"nickname" binding:"max=64"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"max=32"`
	Role     string `json:"role" binding:"required,oneof=admin operator viewer"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Nickname string `json:"nickname" binding:"max=64"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"max=32"`
	Role     string `json:"role" binding:"omitempty,oneof=admin operator viewer"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=64"`
}

// ToggleStatusRequest 切换用户状态请求
type ToggleStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Avatar      string `json:"avatar"`
	Role        string `json:"role"`
	Status      int    `json:"status"`
	MFAEnabled  bool   `json:"mfa_enabled"`
	LastLoginAt string `json:"last_login_at"`
	LastLoginIP string `json:"last_login_ip"`
	CreatedAt   string `json:"created_at"`
}
