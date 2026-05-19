package middleware

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/service"
)

const maxBodyLog = 2048 // 最大记录请求体长度

// sensitivePattern 敏感字段正则（脱敏）
var sensitivePattern = regexp.MustCompile(`(?i)"(password|secret|access_key_secret|private_key|mfa_secret|captcha|token)":\s*"[^"]*"`)

// responseWriter 包装 gin.ResponseWriter 以捕获响应码
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Audit 审计日志中间件
func Audit(auditSvc *service.AuditLogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 读取请求体（跳过文件上传，避免截断 multipart body）
		var bodyStr string
		ct := c.Request.Header.Get("Content-Type")
		isMultipart := strings.HasPrefix(ct, "multipart/form-data")
		if c.Request.Body != nil && c.Request.ContentLength > 0 && !isMultipart {
			bodyBytes, _ := io.ReadAll(io.LimitReader(c.Request.Body, maxBodyLog))
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			bodyStr = sanitizeBody(string(bodyBytes))
		}

		// 包装 ResponseWriter
		w := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = w

		c.Next()

		// 跳过 GET 请求（除了登录和关键查询）和健康检查
		if c.Request.Method == "GET" && !isAuditableGet(c.Request.URL.Path) {
			return
		}
		// 跳过验证码接口
		if strings.HasSuffix(c.Request.URL.Path, "/captcha") {
			return
		}

		// 提取用户信息
		var userID uint
		var username string
		if uid, exists := c.Get("user_id"); exists {
			userID = uid.(uint)
		}
		if uname, exists := c.Get("username"); exists {
			username = uname.(string)
		}

		// 解析模块和操作
		module, action := parseModuleAction(c.Request.Method, c.Request.URL.Path)

		duration := time.Since(start).Milliseconds()

		auditSvc.Record(model.AuditLog{
			UserID:       userID,
			Username:     username,
			Module:       module,
			Action:       action,
			ClientIP:     c.ClientIP(),
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			RequestBody:  bodyStr,
			ResponseCode: c.Writer.Status(),
			Duration:     duration,
		})
	}
}

// sanitizeBody 脱敏请求体中的敏感字段
func sanitizeBody(body string) string {
	return sensitivePattern.ReplaceAllStringFunc(body, func(match string) string {
		idx := strings.Index(match, ":")
		key := match[:idx]
		return key + `: "******"`
	})
}

// parseModuleAction 从请求路径和方法推断模块和操作
func parseModuleAction(method, path string) (module, action string) {
	// 移除 /api/v1/ 前缀
	path = strings.TrimPrefix(path, "/api/v1/")
	parts := strings.Split(path, "/")

	if len(parts) > 0 {
		switch parts[0] {
		case "auth":
			module = "auth"
		case "users":
			module = "user"
		case "assets":
			module = "asset"
		case "cloud-accounts":
			module = "cloud"
		case "sshkeys":
			module = "ssh"
		case "probe":
			module = "probe"
		case "tasks":
			module = "task"
		case "files":
			module = "file"
		case "terminal":
			module = "terminal"
		case "audit":
			module = "audit"
		case "ipam":
			module = "ipam"
		case "settings":
			module = "settings"
		case "linux-users":
			module = "user"
		default:
			module = parts[0]
		}
	}

	switch method {
	case "POST":
		action = "create"
		if strings.Contains(path, "login") {
			action = "login"
		} else if strings.Contains(path, "execute") || strings.Contains(path, "probe") {
			action = "execute"
		} else if strings.Contains(path, "deploy") {
			action = "deploy"
		} else if strings.Contains(path, "sync") {
			action = "sync"
		} else if strings.Contains(path, "distribute") {
			action = "distribute"
		}
	case "PUT":
		action = "update"
	case "DELETE":
		action = "delete"
	default:
		action = "query"
	}

	return
}

// isAuditableGet 判断 GET 请求是否需要审计
func isAuditableGet(path string) bool {
	// 只审计敏感的 GET 操作
	return strings.Contains(path, "/mfa/setup") ||
		strings.Contains(path, "/terminal/sessions") && strings.Contains(path, "/recording")
}
