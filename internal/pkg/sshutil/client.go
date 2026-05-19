package sshutil

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// ClientConfig SSH 客户端配置
type ClientConfig struct {
	Host       string
	Port       int
	User       string
	Password   string
	PrivateKey string
	Timeout    time.Duration
}

// Dial 创建 SSH 连接
func Dial(cfg ClientConfig) (*ssh.Client, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.Port == 0 {
		cfg.Port = 22
	}

	authMethods := buildAuthMethods(cfg)
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no auth method provided")
	}

	sshConfig := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         cfg.Timeout,
	}

	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("ssh dial %s: %w", addr, err)
	}

	return client, nil
}

// buildAuthMethods 构建认证方式列表
func buildAuthMethods(cfg ClientConfig) []ssh.AuthMethod {
	var methods []ssh.AuthMethod

	// 优先使用密钥认证
	if cfg.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(cfg.PrivateKey))
		if err == nil {
			methods = append(methods, ssh.PublicKeys(signer))
		}
	}

	// 密码认证
	if cfg.Password != "" {
		methods = append(methods, ssh.Password(cfg.Password))
	}

	return methods
}
