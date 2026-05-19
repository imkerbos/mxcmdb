package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/pkg/crypto"
	"github.com/imkerbos/mxcmdb/internal/pkg/logger"
	"github.com/imkerbos/mxcmdb/internal/pkg/sshutil"
	"github.com/imkerbos/mxcmdb/internal/repository"
)

// ProbeService 资产探针服务
type ProbeService struct {
	probeRepo *repository.ProbeResultRepository
	assetRepo *repository.AssetRepository
	masterKey string
	configSvc *SystemConfigService
	notifySvc *NotificationService
}

// SetNotificationService 注入通知服务
func (s *ProbeService) SetNotificationService(svc *NotificationService) {
	s.notifySvc = svc
}

// NewProbeService 创建 ProbeService
func NewProbeService(probeRepo *repository.ProbeResultRepository, assetRepo *repository.AssetRepository, masterKey string, configSvc *SystemConfigService) *ProbeService {
	return &ProbeService{
		probeRepo: probeRepo,
		assetRepo: assetRepo,
		masterKey: masterKey,
		configSvc: configSvc,
	}
}

// ProbeAssets 对资产执行探针采集
func (s *ProbeService) ProbeAssets(assetIDs []uint) error {
	timeout := time.Duration(s.configSvc.GetInt("ssh.timeout", 30)) * time.Second
	maxConcurrent := s.configSvc.GetInt("ssh.max_concurrent", 100)
	pool := sshutil.NewWorkerPool(maxConcurrent)

	type assetTask struct {
		task  sshutil.Task
		asset *model.Asset
	}

	// 批量加载资产
	assetList, err := s.assetRepo.GetByIDs(assetIDs)
	if err != nil {
		return fmt.Errorf("加载资产失败: %w", err)
	}
	assetMap := make(map[uint]*model.Asset, len(assetList))
	for i := range assetList {
		assetMap[assetList[i].ID] = &assetList[i]
	}

	var assetTasks []assetTask
	var tasks []sshutil.Task
	for _, id := range assetIDs {
		asset, ok := assetMap[id]
		if !ok {
			continue
		}
		password := ""
		if asset.SshPassword != "" {
			password, _ = crypto.Decrypt(asset.SshPassword, s.masterKey)
		}
		t := sshutil.Task{
			Host: asset.IP, Port: asset.Port, User: asset.SshUser,
			Password: password,
		}
		tasks = append(tasks, t)
		assetTasks = append(assetTasks, assetTask{task: t, asset: asset})
	}

	// 采集命令
	probeCmd := strings.Join([]string{
		`echo "===HOSTNAME==="; hostname`,
		`echo "===CPU==="; lscpu 2>/dev/null || cat /proc/cpuinfo | head -20`,
		`echo "===MEMORY==="; free -m`,
		`echo "===DISK==="; df -h`,
		`echo "===NETWORK==="; ip addr 2>/dev/null || ifconfig`,
		`echo "===OS==="; cat /etc/os-release 2>/dev/null || cat /etc/redhat-release 2>/dev/null`,
		`echo "===KERNEL==="; uname -r`,
		`echo "===DOCKER==="; docker version --format '{{.Server.Version}}' 2>/dev/null || echo "not installed"`,
		`echo "===SERVICES==="; systemctl list-units --type=service --state=running --no-pager --no-legend 2>/dev/null`,
		`echo "===SSHUSERS==="; cat /etc/passwd | grep -v nologin | grep -v /bin/false`,
		`echo "===UPTIME==="; uptime -s 2>/dev/null || awk '{d=int($1/86400);h=int($1%86400/3600);m=int($1%3600/60);printf "%dd %dh %dm\n",d,h,m}' /proc/uptime`,
		`echo "===DNS==="; grep nameserver /etc/resolv.conf 2>/dev/null | awk '{print $2}'`,
		`echo "===GATEWAY==="; ip route 2>/dev/null | grep default || route -n 2>/dev/null | grep UG`,
		`echo "===MANUFACTURER==="; cat /sys/class/dmi/id/sys_vendor 2>/dev/null || echo "unknown"`,
		`echo "===PRODUCT==="; cat /sys/class/dmi/id/product_name 2>/dev/null || echo "unknown"`,
		`echo "===SERIAL==="; cat /sys/class/dmi/id/product_serial 2>/dev/null || echo "unknown"`,
		`echo "===PUBLICIP==="; curl -s --connect-timeout 3 -m 5 ifconfig.me 2>/dev/null || echo ""`,
		`echo "===PROCESSES==="; ps -eo pid,ppid,user,%cpu,%mem,rss,stat,start_time,comm --sort=-%mem 2>/dev/null | head -51`,
		`echo "===LISTENERS==="; ss -tlnp 2>/dev/null | tail -n +2`,
	}, "; ")

	// 异步执行
	go func() {
		results := pool.Execute(tasks, func(task sshutil.Task) sshutil.ExecResult {
			client, err := sshutil.Dial(sshutil.ClientConfig{
				Host: task.Host, Port: task.Port, User: task.User,
				Password: task.Password, Timeout: timeout,
			})
			if err != nil {
				return sshutil.ExecResult{Err: err}
			}
			defer client.Close()
			return sshutil.RunCommand(client, probeCmd, timeout*2)
		})

		now := time.Now()
		for i, r := range results {
			if i >= len(assetTasks) {
				break
			}
			asset := assetTasks[i].asset
			probe := &model.ProbeResult{
				AssetID:     asset.ID,
				CollectedAt: now,
			}

			if r.Result.Err != nil {
				probe.Status = "failed"
				probe.ErrorMessage = r.Result.Err.Error()
			} else {
				probe.Status = "success"
				parseProbeOutput(r.Result.Stdout, probe)
			}

			if err := s.probeRepo.Create(probe); err != nil {
				logger.Log.Errorf("保存探针结果失败: %v", err)
			}

			// 更新资产探测时间和状态
			asset.ProbeLastAt = &now
			if asset.Source == "manual" {
				if probe.Status == "success" {
					asset.Status = "online"
				} else {
					asset.Status = "offline"
				}
			}
			if probe.Hostname != "" {
				asset.Hostname = probe.Hostname
			}
			if probe.OS != "" {
				asset.OS = probe.OS
			}
			_ = s.assetRepo.Update(asset)
		}
		logger.Log.Infof("探针采集完成，%d 台主机", len(results))

		// 发送通知
		if s.notifySvc != nil {
			successCount := 0
			failCount := 0
			for _, r := range results {
				if r.Result.Err == nil {
					successCount++
				} else {
					failCount++
				}
			}
			nType := "success"
			if failCount > 0 && successCount == 0 {
				nType = "error"
			} else if failCount > 0 {
				nType = "warning"
			}
			title := "探针采集完成"
			content := ""
			if failCount > 0 {
				content = fmt.Sprintf("成功 %d 台，失败 %d 台", successCount, failCount)
			} else {
				content = fmt.Sprintf("全部 %d 台采集成功", successCount)
			}
			s.notifySvc.NotifyAllOperators(title, content, nType, "probe")
		}
	}()

	return nil
}

// GetLatestResult 获取最新探针结果
func (s *ProbeService) GetLatestResult(assetID uint) (*model.ProbeResult, error) {
	return s.probeRepo.GetLatestByAssetID(assetID)
}

// ListResults 查询探针历史
func (s *ProbeService) ListResults(assetID uint, limit int) ([]model.ProbeResult, error) {
	return s.probeRepo.ListByAssetID(assetID, limit)
}

// parseProbeOutput 解析探针输出
func parseProbeOutput(output string, probe *model.ProbeResult) {
	sections := map[string]string{}
	currentSection := ""
	var currentLines []string

	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "===") && strings.HasSuffix(line, "===") {
			if currentSection != "" {
				sections[currentSection] = strings.TrimSpace(strings.Join(currentLines, "\n"))
			}
			currentSection = strings.Trim(line, "=")
			currentLines = nil
		} else {
			currentLines = append(currentLines, line)
		}
	}
	if currentSection != "" {
		sections[currentSection] = strings.TrimSpace(strings.Join(currentLines, "\n"))
	}

	probe.Hostname = sections["HOSTNAME"]
	probe.CPU = sections["CPU"]
	probe.Memory = sections["MEMORY"]
	probe.Disk = sections["DISK"]
	probe.Network = sections["NETWORK"]
	probe.OS = extractPrettyName(sections["OS"])
	probe.Kernel = sections["KERNEL"]
	probe.DockerVersion = sections["DOCKER"]
	probe.RunningServices = sections["SERVICES"]
	probe.SSHUsers = sections["SSHUSERS"]
	probe.Uptime = sections["UPTIME"]
	probe.DNS = sections["DNS"]
	probe.Gateway = sections["GATEWAY"]
	probe.Manufacturer = sanitizeDMI(sections["MANUFACTURER"])
	probe.ProductModel = sanitizeDMI(sections["PRODUCT"])
	probe.SerialNumber = sanitizeDMI(sections["SERIAL"])
	probe.PublicIP = strings.TrimSpace(sections["PUBLICIP"])
	probe.Processes = sections["PROCESSES"]
	probe.Listeners = sections["LISTENERS"]
}

// extractPrettyName 从 /etc/os-release 内容中提取 PRETTY_NAME
func extractPrettyName(osRelease string) string {
	for _, line := range strings.Split(osRelease, "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			val := strings.TrimPrefix(line, "PRETTY_NAME=")
			return strings.Trim(val, "\"")
		}
	}
	// 非 os-release 格式（如 /etc/redhat-release），直接截断返回
	if len(osRelease) > 128 {
		return osRelease[:128]
	}
	return osRelease
}

// sanitizeDMI 清理 DMI 信息，过滤无意义值
func sanitizeDMI(val string) string {
	v := strings.TrimSpace(val)
	lower := strings.ToLower(v)
	if lower == "unknown" || lower == "not specified" || lower == "to be filled by o.e.m." || lower == "default string" || lower == "" {
		return ""
	}
	return v
}
