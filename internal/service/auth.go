package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务
type AuthService struct {
	userRepo   *repository.UserRepository
	mfaSvc     *MFAService
	configSvc  *SystemConfigService
	jwtSecret  string
	accessExp  time.Duration
	refreshExp time.Duration
}

// NewAuthService 创建 AuthService
func NewAuthService(userRepo *repository.UserRepository, mfaSvc *MFAService, configSvc *SystemConfigService, jwtSecret string, accessExp, refreshExp time.Duration) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		mfaSvc:     mfaSvc,
		configSvc:  configSvc,
		jwtSecret:  jwtSecret,
		accessExp:  accessExp,
		refreshExp: refreshExp,
	}
}

// Login 用户登录（密码验证 + MFA 校验 + MFA 策略强制）
func (s *AuthService) Login(username, password, mfaCode string) (*dto.LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	if user.Status != 1 {
		return nil, errors.New("用户已被禁用")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// MFA 校验
	if user.MFAEnabled {
		if mfaCode == "" {
			return &dto.LoginResponse{RequireMFA: true}, nil
		}
		if !s.mfaSvc.Verify(user, mfaCode) {
			return nil, errors.New("MFA 验证码错误")
		}
	}

	// MFA 策略强制检查：用户已通过密码验证但未绑定 MFA
	if !user.MFAEnabled && s.isMFARequired(user) {
		resp, err := s.issueTokens(user)
		if err != nil {
			return nil, err
		}
		resp.RequireMFASetup = true
		return resp, nil
	}

	return s.issueTokens(user)
}

// isMFARequired 判断当前用户是否需要强制开启 MFA
func (s *AuthService) isMFARequired(user *model.User) bool {
	policy := s.configSvc.GetWithDefault("security.mfa_policy", "optional")
	switch policy {
	case "required_all":
		return true
	case "required_admin":
		return user.Role == "admin"
	default:
		return false
	}
}

// GetUserByID 根据 ID 获取用户模型
func (s *AuthService) GetUserByID(userID uint) (*model.User, error) {
	return s.userRepo.GetByID(userID)
}

// GetUserInfo 获取用户信息
func (s *AuthService) GetUserInfo(userID uint) (*dto.UserInfo, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	return &dto.UserInfo{
		ID:         user.ID,
		Username:   user.Username,
		Nickname:   user.Nickname,
		Email:      user.Email,
		Phone:      user.Phone,
		Role:       user.Role,
		MFAEnabled: user.MFAEnabled,
	}, nil
}

func (s *AuthService) issueTokens(user *model.User) (*dto.LoginResponse, error) {
	accessToken, err := s.generateToken(user.ID, user.Username, user.Role, s.accessExp)
	if err != nil {
		return nil, errors.New("生成 Token 失败")
	}

	refreshToken, err := s.generateToken(user.ID, user.Username, user.Role, s.refreshExp)
	if err != nil {
		return nil, errors.New("生成 Token 失败")
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessExp.Seconds()),
	}, nil
}

func (s *AuthService) generateToken(userID uint, username, role string, expiration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(expiration).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
