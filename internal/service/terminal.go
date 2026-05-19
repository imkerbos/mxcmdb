package service

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/pkg/crypto"
	"github.com/imkerbos/mxcmdb/internal/pkg/logger"
	"github.com/imkerbos/mxcmdb/internal/pkg/sshutil"
	"github.com/imkerbos/mxcmdb/internal/repository"
	"golang.org/x/crypto/ssh"
)

// activeSession 活跃会话内部记录
type activeSession struct {
	sessionID  uint
	assetID    uint
	hostname   string
	ip         string
	userID     uint
	username   string
	clientIP   string
	startedAt  time.Time
	wsConn     *websocket.Conn
	sshSession *ssh.Session
}

// TerminalService Web Terminal 服务
type TerminalService struct {
	sessionRepo    *repository.TerminalSessionRepository
	assetRepo      *repository.AssetRepository
	masterKey      string
	configSvc      *SystemConfigService
	activeSessions sync.Map // map[uint]*activeSession，key 为 DB session ID
}

// NewTerminalService 创建 TerminalService
func NewTerminalService(sessionRepo *repository.TerminalSessionRepository, assetRepo *repository.AssetRepository, masterKey string, configSvc *SystemConfigService) *TerminalService {
	return &TerminalService{
		sessionRepo: sessionRepo,
		assetRepo:   assetRepo,
		masterKey:   masterKey,
		configSvc:   configSvc,
	}
}

// CleanupStaleSessions 清理上次进程残留的僵尸会话
func (s *TerminalService) CleanupStaleSessions() {
	count, err := s.sessionRepo.CleanupStaleSessions()
	if err != nil {
		logger.Log.Errorf("清理僵尸终端会话失败: %v", err)
		return
	}
	if count > 0 {
		logger.Log.Infof("已清理 %d 个僵尸终端会话", count)
	}
}

// HandleWebSocket 处理终端 WebSocket 连接
func (s *TerminalService) HandleWebSocket(conn *websocket.Conn, assetID, userID uint, username, clientIP string) {
	asset, err := s.assetRepo.GetByID(assetID)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\r\n\033[31m✗ 资产不存在: %v\033[0m\r\n", err)))
		return
	}

	// 解密 SSH 密码
	password := ""
	if asset.SshPassword != "" {
		password, _ = crypto.Decrypt(asset.SshPassword, s.masterKey)
	}

	timeout := time.Duration(s.configSvc.GetInt("ssh.timeout", 30)) * time.Second

	// 连接状态提示
	_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(
		"\033[33m⟳ 正在连接 %s@%s:%d ...\033[0m\r\n", asset.SshUser, asset.IP, asset.Port)))

	// 建立 SSH 连接
	sshClient, err := sshutil.Dial(sshutil.ClientConfig{
		Host:     asset.IP,
		Port:     asset.Port,
		User:     asset.SshUser,
		Password: password,
		Timeout:  timeout,
	})
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\r\n\033[31m✗ SSH 连接失败: %v\033[0m\r\n", err)))
		return
	}
	defer sshClient.Close()

	_ = conn.WriteMessage(websocket.TextMessage, []byte("\033[33m⟳ 正在创建会话 ...\033[0m\r\n"))

	// 创建 SSH Session
	session, err := sshClient.NewSession()
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\r\n\033[31m✗ 创建会话失败: %v\033[0m\r\n", err)))
		return
	}
	defer session.Close()

	// 请求 PTY
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", 40, 120, modes); err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\r\n\033[31m✗ 请求 PTY 失败: %v\033[0m\r\n", err)))
		return
	}

	// 获取 stdin/stdout
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return
	}
	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return
	}

	// 启动 shell
	if err := session.Shell(); err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\r\n\033[31m✗ 启动 Shell 失败: %v\033[0m\r\n", err)))
		return
	}

	// 连接成功提示
	_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(
		"\033[32m✓ 已连接 %s@%s\033[0m\r\n\r\n", asset.SshUser, asset.Hostname)))

	// 创建会话记录
	now := time.Now()
	termSession := &model.TerminalSession{
		AssetID:   assetID,
		UserID:    userID,
		Username:  username,
		Status:    "connected",
		ClientIP:  clientIP,
		StartedAt: now,
	}
	_ = s.sessionRepo.Create(termSession)

	// 注册活跃会话
	s.activeSessions.Store(termSession.ID, &activeSession{
		sessionID:  termSession.ID,
		assetID:    assetID,
		hostname:   asset.Hostname,
		ip:         asset.IP,
		userID:     userID,
		username:   username,
		clientIP:   clientIP,
		startedAt:  now,
		wsConn:     conn,
		sshSession: session,
	})

	// asciicast v2 录制
	var recording strings.Builder
	recordStart := time.Now()
	header, _ := json.Marshal(map[string]any{
		"version":   2,
		"width":     120,
		"height":    40,
		"timestamp": recordStart.Unix(),
		"env":       map[string]string{"SHELL": "/bin/bash", "TERM": "xterm-256color"},
	})
	recording.Write(header)
	recording.WriteByte('\n')
	var recordMu sync.Mutex

	appendRecord := func(data string) {
		recordMu.Lock()
		defer recordMu.Unlock()
		elapsed := time.Since(recordStart).Seconds()
		entry, _ := json.Marshal([]any{elapsed, "o", data})
		recording.Write(entry)
		recording.WriteByte('\n')
	}

	done := make(chan struct{})

	// SSH stdout -> WebSocket
	go func() {
		defer close(done)
		buf := make([]byte, 8192)
		for {
			n, err := stdoutPipe.Read(buf)
			if err != nil {
				break
			}
			data := string(buf[:n])
			appendRecord(data)
			if err := conn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				break
			}
		}
	}()

	// WebSocket -> SSH stdin
	go func() {
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				session.Close()
				break
			}
			if msgType == websocket.TextMessage {
				// 处理终端窗口大小调整
				if len(msg) > 0 && msg[0] == '{' {
					var resize struct {
						Cols int `json:"cols"`
						Rows int `json:"rows"`
					}
					if json.Unmarshal(msg, &resize) == nil && resize.Cols > 0 && resize.Rows > 0 {
						_ = session.WindowChange(resize.Rows, resize.Cols)
						continue
					}
				}
				_, _ = stdinPipe.Write(msg)
			}
		}
	}()

	// 等待连接结束
	<-done
	_ = session.Wait()

	// 注销活跃会话
	s.activeSessions.Delete(termSession.ID)

	// 保存会话
	finishNow := time.Now()
	termSession.Status = "disconnected"
	termSession.FinishedAt = &finishNow
	termSession.Recording = compressRecording(recording.String())
	_ = s.sessionRepo.Update(termSession)

	logger.Log.Infof("终端会话 %d 结束: user=%s, asset=%s", termSession.ID, username, asset.IP)
}

// ListSessions 会话列表
func (s *TerminalService) ListSessions(page, pageSize int, status string) ([]dto.TerminalSessionResponse, int64, error) {
	sessions, total, err := s.sessionRepo.List(page, pageSize, status)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.TerminalSessionResponse, len(sessions))
	for i, sess := range sessions {
		result[i] = dto.TerminalSessionResponse{
			ID:        sess.ID,
			AssetID:   sess.AssetID,
			Hostname:  sess.Asset.Hostname,
			IP:        sess.Asset.IP,
			UserID:    sess.UserID,
			Username:  sess.Username,
			Status:    sess.Status,
			ClientIP:  sess.ClientIP,
			StartedAt: sess.StartedAt.Format("2006-01-02 15:04:05"),
		}
		if sess.FinishedAt != nil {
			result[i].FinishedAt = sess.FinishedAt.Format("2006-01-02 15:04:05")
		}
	}
	return result, total, nil
}

// GetRecording 获取会话录制
func (s *TerminalService) GetRecording(id uint) (string, error) {
	session, err := s.sessionRepo.GetByID(id)
	if err != nil {
		return "", err
	}
	return decompressRecording(session.Recording), nil
}

// ListActiveSessions 获取在线会话列表
func (s *TerminalService) ListActiveSessions() []dto.ActiveSessionResponse {
	var result []dto.ActiveSessionResponse
	s.activeSessions.Range(func(_, value any) bool {
		as := value.(*activeSession)
		dur := time.Since(as.startedAt)
		h := int(dur.Hours())
		m := int(dur.Minutes()) % 60
		sec := int(dur.Seconds()) % 60
		result = append(result, dto.ActiveSessionResponse{
			SessionID: as.sessionID,
			AssetID:   as.assetID,
			Hostname:  as.hostname,
			IP:        as.ip,
			UserID:    as.userID,
			Username:  as.username,
			ClientIP:  as.clientIP,
			StartedAt: as.startedAt.Format("2006-01-02 15:04:05"),
			Duration:  fmt.Sprintf("%02d:%02d:%02d", h, m, sec),
		})
		return true
	})
	return result
}

// compressRecording 使用 gzip 压缩录制数据并 base64 编码
func compressRecording(raw string) string {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write([]byte(raw))
	_ = gz.Close()
	return "gz:" + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// decompressRecording 解压录制数据
func decompressRecording(data string) string {
	if !strings.HasPrefix(data, "gz:") {
		return data // 未压缩的旧数据，直接返回
	}
	decoded, err := base64.StdEncoding.DecodeString(data[3:])
	if err != nil {
		return data
	}
	gz, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return data
	}
	defer gz.Close()
	out, err := io.ReadAll(gz)
	if err != nil {
		return data
	}
	return string(out)
}

// KillSession 强制终止会话
func (s *TerminalService) KillSession(id uint) error {
	val, ok := s.activeSessions.Load(id)
	if !ok {
		return fmt.Errorf("会话 %d 不在线", id)
	}
	as := val.(*activeSession)
	// 关闭 SSH session 和 WebSocket，HandleWebSocket 会自动完成清理
	_ = as.sshSession.Close()
	_ = as.wsConn.Close()
	logger.Log.Infof("终端会话 %d 被强制终止: user=%s, asset=%s", id, as.username, as.ip)
	return nil
}
