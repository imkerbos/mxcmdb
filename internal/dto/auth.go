package dto

// LoginRequest 登录请求
type LoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	CaptchaID string `json:"captcha_id" binding:"required"`
	Captcha   string `json:"captcha" binding:"required"`
	MFACode   string `json:"mfa_code"` // MFA 验证码（启用 MFA 时必填）
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token"`
	ExpiresIn       int64  `json:"expires_in"`
	RequireMFA      bool   `json:"require_mfa"`       // 需要 MFA 二次验证
	RequireMFASetup bool   `json:"require_mfa_setup"` // 策略要求绑定 MFA（已登录但需跳转绑定页）
}

// MFASetupRequest MFA 绑定请求
type MFASetupRequest struct {
	Code string `json:"code" binding:"required"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Role       string `json:"role"`
	MFAEnabled bool   `json:"mfa_enabled"`
}
