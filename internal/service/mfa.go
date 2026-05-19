package service

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/repository"
)

// MFAService MFA 服务
type MFAService struct {
	userRepo *repository.UserRepository
	issuer   string
}

// NewMFAService 创建 MFAService
func NewMFAService(userRepo *repository.UserRepository, issuer string) *MFAService {
	return &MFAService{
		userRepo: userRepo,
		issuer:   issuer,
	}
}

// MFASetupResult MFA 绑定结果
type MFASetupResult struct {
	Secret string `json:"secret"`
	URL    string `json:"url"`
}

// Setup 生成 MFA 密钥和 otpauth URL
func (s *MFAService) Setup(user *model.User) (*MFASetupResult, error) {
	secret := generateSecret(16)
	url := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
		s.issuer, user.Username, secret, s.issuer)

	return &MFASetupResult{
		Secret: secret,
		URL:    url,
	}, nil
}

// Bind 绑定 MFA（验证一次 code 后写入用户记录）
func (s *MFAService) Bind(userID uint, secret, code string) error {
	if !verifyTOTP(secret, code) {
		return errors.New("验证码错误，请重试")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	user.MFASecret = secret
	user.MFAEnabled = true
	return s.userRepo.Update(user)
}

// Verify 校验 MFA 验证码
func (s *MFAService) Verify(user *model.User, code string) bool {
	if !user.MFAEnabled || user.MFASecret == "" {
		return true // 未启用 MFA 则跳过
	}
	return verifyTOTP(user.MFASecret, code)
}

// generateSecret 生成 base32 编码的随机密钥
func generateSecret(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = byte(rand.Intn(256))
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(b), "=")
}

// verifyTOTP 校验 TOTP（允许前后 30 秒窗口）
func verifyTOTP(secret, code string) bool {
	now := time.Now().Unix() / 30
	for _, offset := range []int64{-1, 0, 1} {
		if generateTOTP(secret, now+offset) == code {
			return true
		}
	}
	return false
}

// generateTOTP 根据密钥和时间步生成 6 位 TOTP
func generateTOTP(secret string, counter int64) string {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return ""
	}

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	hash := mac.Sum(nil)

	offset := hash[len(hash)-1] & 0x0f
	code := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	return fmt.Sprintf("%06d", code%1000000)
}
