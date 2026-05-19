package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/pkg/response"
	"github.com/imkerbos/mxcmdb/internal/service"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authSvc    *service.AuthService
	captchaSvc *service.CaptchaService
	mfaSvc     *service.MFAService
}

// NewAuthHandler 创建 AuthHandler
func NewAuthHandler(authSvc *service.AuthService, captchaSvc *service.CaptchaService, mfaSvc *service.MFAService) *AuthHandler {
	return &AuthHandler{
		authSvc:    authSvc,
		captchaSvc: captchaSvc,
		mfaSvc:     mfaSvc,
	}
}

// GetCaptcha 获取验证码
func (h *AuthHandler) GetCaptcha(c *gin.Context) {
	result, err := h.captchaSvc.Generate(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "生成验证码失败")
		return
	}
	response.Success(c, result)
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	// 校验验证码
	if !h.captchaSvc.Verify(c.Request.Context(), req.CaptchaID, req.Captcha) {
		response.Error(c, http.StatusBadRequest, 40002, "验证码错误")
		return
	}

	resp, err := h.authSvc.Login(req.Username, req.Password, req.MFACode)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, 40100, err.Error())
		return
	}

	response.Success(c, resp)
}

// GetUserInfo 获取当前用户信息
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID, _ := c.Get("user_id")
	info, err := h.authSvc.GetUserInfo(userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "获取用户信息失败")
		return
	}
	response.Success(c, info)
}

// SetupMFA 生成 MFA 密钥
func (h *AuthHandler) SetupMFA(c *gin.Context) {
	userID, _ := c.Get("user_id")
	user, err := h.authSvc.GetUserByID(userID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "用户不存在")
		return
	}

	result, err := h.mfaSvc.Setup(user)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 50001, "生成 MFA 密钥失败")
		return
	}
	response.Success(c, result)
}

// BindMFA 绑定 MFA
func (h *AuthHandler) BindMFA(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req dto.MFASetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	// 从请求中获取 secret（前端在 setup 时保存的）
	secret := c.Query("secret")
	if secret == "" {
		response.Error(c, http.StatusBadRequest, 40001, "缺少 MFA secret")
		return
	}

	if err := h.mfaSvc.Bind(userID.(uint), secret, req.Code); err != nil {
		response.Error(c, http.StatusBadRequest, 40003, err.Error())
		return
	}
	response.Success(c, nil)
}
