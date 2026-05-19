package service

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/pkg/crypto"
	"github.com/imkerbos/mxcmdb/internal/pkg/logger"
	"github.com/imkerbos/mxcmdb/internal/repository"
	"golang.org/x/crypto/ssh"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// ===================================================================
//  终端录制端到端测试
//  Mock SSH Server + 真实 WebSocket → 验证完整录制链路
// ===================================================================

func getDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost port=5432 user=mxcmdb password=mxcmdb123 dbname=mxcmdb sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Skipf("跳过（数据库不可用）: %v", err)
	}
	return db
}

// --- Mock SSH Server ---

func startMockSSHServer(t *testing.T) (addr string, cleanup func()) {
	key, err := ssh.ParsePrivateKey([]byte(testHostKey))
	if err != nil {
		t.Fatalf("解析 host key: %v", err)
	}

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			return nil, nil // 接受任意密码
		},
	}
	config.AddHostKey(key)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("监听失败: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleSSHConn(conn, config)
		}
	}()

	return listener.Addr().String(), func() {
		listener.Close()
		<-done
	}
}

func handleSSHConn(nConn net.Conn, config *ssh.ServerConfig) {
	defer nConn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		return
	}
	defer sshConn.Close()
	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}
		go func() {
			for req := range requests {
				switch req.Type {
				case "pty-req", "env":
					if req.WantReply {
						req.Reply(true, nil)
					}
				case "shell":
					if req.WantReply {
						req.Reply(true, nil)
					}
					// 模拟终端交互
					fmt.Fprintf(channel, "\033[32mroot@mock-host\033[0m:~# ")
					time.Sleep(80 * time.Millisecond)
					fmt.Fprintf(channel, "hostname\r\nmock-host\r\n")
					time.Sleep(80 * time.Millisecond)
					fmt.Fprintf(channel, "\033[32mroot@mock-host\033[0m:~# uptime\r\n")
					time.Sleep(80 * time.Millisecond)
					fmt.Fprintf(channel, " 10:00:00 up 30 days, load average: 0.01, 0.05, 0.10\r\n")
					time.Sleep(50 * time.Millisecond)
					channel.Close()
				case "window-change":
					if req.WantReply {
						req.Reply(true, nil)
					}
				default:
					if req.WantReply {
						req.Reply(false, nil)
					}
				}
			}
		}()
	}
}

// TestTerminalRecording_E2E 完整端到端:
// Mock SSH → WebSocket 连接 → 终端交互 → 录制 → 压缩存储 → 解压读取 → 格式校验
func TestTerminalRecording_E2E(t *testing.T) {
	// 初始化全局 logger（测试环境）
	if logger.Log == nil {
		_ = logger.Init("info", "text")
	}

	db := getDB(t)

	// 1. 启动 Mock SSH
	sshAddr, sshCleanup := startMockSSHServer(t)
	defer sshCleanup()
	host, portStr, _ := net.SplitHostPort(sshAddr)
	var sshPort int
	fmt.Sscanf(portStr, "%d", &sshPort)
	t.Logf("Step 1: Mock SSH 启动 → %s", sshAddr)

	// 2. 创建测试资产（加密一个测试密码）
	testMasterKey := ""
	encryptedPwd, err := crypto.Encrypt("testpass", testMasterKey)
	if err != nil {
		t.Fatalf("加密测试密码失败: %v", err)
	}
	asset := &model.Asset{
		Hostname:    "e2e-mock-host",
		IP:          host,
		Port:        sshPort,
		SshUser:     "root",
		SshPassword: encryptedPwd,
		Type:        "server",
		Source:      "manual",
		Status:      "unknown",
	}
	db.Create(asset)
	defer db.Unscoped().Delete(&model.Asset{}, asset.ID)
	t.Logf("Step 2: 测试资产创建 → id=%d, %s:%d", asset.ID, host, sshPort)

	// 3. 构建服务 + 路由
	sessionRepo := repository.NewTerminalSessionRepository(db)
	assetRepo := repository.NewAssetRepository(db)
	configRepo := repository.NewSystemConfigRepository(db)
	configSvc := NewSystemConfigService(configRepo)
	_ = configSvc.LoadAll()
	termSvc := NewTerminalService(sessionRepo, assetRepo, testMasterKey, configSvc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ws/terminal/:assetId", func(c *gin.Context) {
		// 直接注入 user 上下文（绕过 JWT）
		c.Set("user_id", uint(1))
		c.Set("username", "test-admin")

		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		termSvc.HandleWebSocket(conn, asset.ID, 1, "test-admin", "127.0.0.1")
	})

	ts := httptest.NewServer(r)
	defer ts.Close()
	t.Logf("Step 3: HTTP 测试服务启动 → %s", ts.URL)

	// 4. WebSocket 连接
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + fmt.Sprintf("/ws/terminal/%d", asset.ID)
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket 连接失败: %v", err)
	}
	t.Log("Step 4: WebSocket 连接成功")

	// 5. 读取终端输出
	var received strings.Builder
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		for {
			_, msg, err := wsConn.ReadMessage()
			if err != nil {
				break
			}
			received.Write(msg)
		}
	}()

	// 等待 Mock SSH 自动关闭 + HandleWebSocket 完成
	<-readDone
	wsConn.Close()
	// 等一下让 DB 写入完成
	time.Sleep(200 * time.Millisecond)

	t.Logf("Step 5: 终端输出 %d bytes", received.Len())
	if received.Len() == 0 {
		t.Fatal("没有收到终端输出")
	}

	// 6. 验证终端输出内容
	output := received.String()
	if !strings.Contains(output, "mock-host") {
		t.Errorf("输出中缺少 hostname: %s", output[:min(200, len(output))])
	}
	if !strings.Contains(output, "load average") {
		t.Errorf("输出中缺少 uptime")
	}
	t.Log("Step 6: 终端输出内容验证 PASS")

	// 7. 检查数据库会话
	var sessions []model.TerminalSession
	db.Where("asset_id = ?", asset.ID).Order("id DESC").Find(&sessions)
	defer func() {
		for _, s := range sessions {
			db.Unscoped().Delete(&model.TerminalSession{}, s.ID)
		}
	}()

	if len(sessions) == 0 {
		t.Fatal("数据库中没有会话记录")
	}
	sess := sessions[0]
	t.Logf("Step 7: 数据库会话 → id=%d, status=%s, recording=%d bytes", sess.ID, sess.Status, len(sess.Recording))

	if sess.Status != "disconnected" {
		t.Errorf("会话状态不对: %s (期望 disconnected)", sess.Status)
	}

	// 8. 验证录制数据是压缩格式 (gz: 前缀)
	if !strings.HasPrefix(sess.Recording, "gz:") {
		t.Fatalf("录制数据未压缩！前30字符: [%s]", sess.Recording[:min(30, len(sess.Recording))])
	}
	t.Log("Step 8: 数据库存储格式 → gz 压缩 PASS")

	// 9. 通过 Service 读取（测试解压路径）
	recordingData, err := termSvc.GetRecording(sess.ID)
	if err != nil {
		t.Fatalf("GetRecording 失败: %v", err)
	}
	t.Logf("Step 9: GetRecording 解压 → %d bytes", len(recordingData))

	// 10. 验证 asciicast v2 格式（前端 SessionPlayer 需要的格式）
	lines := strings.Split(strings.TrimSpace(recordingData), "\n")
	if len(lines) < 2 {
		t.Fatalf("录制行数太少: %d", len(lines))
	}

	// Header
	var header map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &header); err != nil {
		t.Fatalf("Header JSON 解析失败: %v\n内容: %s", err, lines[0])
	}
	if header["version"] != float64(2) {
		t.Errorf("version != 2: %v", header["version"])
	}
	if header["width"] != float64(120) {
		t.Errorf("width != 120: %v", header["width"])
	}
	t.Logf("Step 10a: Header → version=%v, width=%v, height=%v", header["version"], header["width"], header["height"])

	// Events
	eventCount := 0
	var lastTime float64
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var event []any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Errorf("事件解析失败: %v", err)
			continue
		}
		if len(event) != 3 {
			t.Errorf("事件格式异常 (len=%d): %v", len(event), event)
			continue
		}
		ts := event[0].(float64)
		if ts < lastTime {
			t.Errorf("时间戳不递增: %.3f < %.3f", ts, lastTime)
		}
		lastTime = ts
		if event[1] != "o" {
			t.Errorf("事件类型不是 'o': %v", event[1])
		}
		eventCount++
	}
	t.Logf("Step 10b: Events → %d 条, 时间跨度 %.3fs", eventCount, lastTime)

	if eventCount == 0 {
		t.Fatal("录制事件为空")
	}

	// 11. 录制数据包含终端交互内容
	if !strings.Contains(recordingData, "mock-host") {
		t.Errorf("录制数据中缺少 mock-host")
	}
	if !strings.Contains(recordingData, "load average") {
		t.Errorf("录制数据中缺少 uptime 输出")
	}
	t.Log("Step 11: 录制内容验证 PASS")

	// 12. 压缩比
	ratio := float64(len(sess.Recording)) / float64(len(recordingData)) * 100
	t.Logf("Step 12: 压缩比 → 原始 %d bytes → 压缩 %d bytes (%.1f%%)", len(recordingData), len(sess.Recording), ratio)

	t.Log("")
	t.Log("========================================")
	t.Log("  端到端测试 ALL PASS")
	t.Log("  Mock SSH → WebSocket → 终端交互")
	t.Log("  → asciicast 录制 → gzip 压缩存储")
	t.Log("  → 解压读取 → 前端格式校验")
	t.Log("========================================")
}

// testHostKey 测试用 Ed25519 私钥（注意：行首不能有缩进）
const testHostKey = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACCj1lFwXb8WjJCMa0hq7JfklGTJZKVl0molALkMal09dQAAAJgeMWe2HjFn
tgAAAAtzc2gtZWQyNTUxOQAAACCj1lFwXb8WjJCMa0hq7JfklGTJZKVl0molALkMal09dQ
AAAEDr56zKZyC2Hi1krbuhIcOPM5FqFWdEwPZfwfgz3DuZ2KPWUXBdvxaMkIxrSGrsl+SU
ZMlkpWXSaiUAuQxqXT11AAAADnRlc3RAbG9jYWxob3N0AQIDBAUGBw==
-----END OPENSSH PRIVATE KEY-----`
