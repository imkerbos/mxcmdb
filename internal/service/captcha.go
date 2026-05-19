package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

const captchaKeyPrefix = "captcha:"

// CaptchaService 验证码服务
type CaptchaService struct {
	rdb    *redis.Client
	length int
	expire time.Duration
}

// NewCaptchaService 创建 CaptchaService
func NewCaptchaService(rdb *redis.Client, length int, expireSeconds int) *CaptchaService {
	return &CaptchaService{
		rdb:    rdb,
		length: length,
		expire: time.Duration(expireSeconds) * time.Second,
	}
}

// CaptchaResult 验证码生成结果
type CaptchaResult struct {
	CaptchaID string `json:"captcha_id"`
	ImageB64  string `json:"captcha_image"`
}

// Generate 生成验证码，返回 captcha_id 和 base64 图片
func (s *CaptchaService) Generate(ctx context.Context) (*CaptchaResult, error) {
	code := s.randomCode()
	captchaID := fmt.Sprintf("%d%s", time.Now().UnixNano(), s.randomCode())

	key := captchaKeyPrefix + captchaID
	if err := s.rdb.Set(ctx, key, code, s.expire).Err(); err != nil {
		return nil, fmt.Errorf("store captcha: %w", err)
	}

	img := generateCaptchaImage(code)

	return &CaptchaResult{
		CaptchaID: captchaID,
		ImageB64:  img,
	}, nil
}

// Verify 校验验证码
func (s *CaptchaService) Verify(ctx context.Context, captchaID, code string) bool {
	key := captchaKeyPrefix + captchaID
	stored, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	// 验证后删除，防止重复使用
	s.rdb.Del(ctx, key)
	return stored == code
}

func (s *CaptchaService) randomCode() string {
	digits := "0123456789"
	b := make([]byte, s.length)
	for i := range b {
		b[i] = digits[rand.Intn(len(digits))]
	}
	return string(b)
}
