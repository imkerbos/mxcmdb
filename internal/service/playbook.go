package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/pkg/crypto"
	"github.com/imkerbos/mxcmdb/internal/pkg/logger"
	"github.com/imkerbos/mxcmdb/internal/pkg/sshutil"
	"github.com/imkerbos/mxcmdb/internal/pkg/ws"
	"github.com/imkerbos/mxcmdb/internal/repository"
	"golang.org/x/crypto/ssh"
)

var playbookJobCounter atomic.Uint64

// PlaybookService 剧本服务
type PlaybookService struct {
	repo      *repository.PlaybookRepository
	assetRepo *repository.AssetRepository
	masterKey string
	configSvc *SystemConfigService
	hub       *ws.Hub
}

// NewPlaybookService 创建 PlaybookService
func NewPlaybookService(
	repo *repository.PlaybookRepository,
	assetRepo *repository.AssetRepository,
	masterKey string,
	configSvc *SystemConfigService,
	hub *ws.Hub,
) *PlaybookService {
	return &PlaybookService{
		repo:      repo,
		assetRepo: assetRepo,
		masterKey: masterKey,
		configSvc: configSvc,
		hub:       hub,
	}
}

// List 查询所有剧本
func (s *PlaybookService) List() ([]dto.PlaybookResponse, error) {
	playbooks, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	var result []dto.PlaybookResponse
	for _, pb := range playbooks {
		result = append(result, s.toResponse(&pb))
	}
	return result, nil
}

// GetByID 获取剧本详情
func (s *PlaybookService) GetByID(id uint) (*dto.PlaybookResponse, error) {
	pb, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("剧本不存在")
	}
	resp := s.toResponse(pb)
	return &resp, nil
}

// Create 创建自定义剧本
func (s *PlaybookService) Create(req dto.CreatePlaybookRequest, userID uint) (*dto.PlaybookResponse, error) {
	pb := &model.Playbook{
		Name:        req.Name,
		Description: req.Description,
		Category:    "custom",
		IsBuiltin:   false,
		Enabled:     true,
		CreatedBy:   userID,
	}

	for i, step := range req.Steps {
		timeout := step.Timeout
		if timeout <= 0 {
			timeout = 60
		}
		onError := step.OnError
		if onError == "" {
			onError = "stop"
		}
		pb.Steps = append(pb.Steps, model.PlaybookStep{
			SortOrder: i,
			Name:      step.Name,
			Script:    step.Script,
			Timeout:   timeout,
			OnError:   onError,
		})
	}

	if err := s.repo.Create(pb); err != nil {
		return nil, fmt.Errorf("创建剧本失败: %w", err)
	}

	resp := s.toResponse(pb)
	return &resp, nil
}

// Update 更新剧本（内置剧本不可修改名称和步骤，只能启用/禁用）
func (s *PlaybookService) Update(id uint, req dto.UpdatePlaybookRequest) error {
	pb, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("剧本不存在")
	}

	if pb.IsBuiltin {
		// 内置剧本只能切换启用状态
		if req.Enabled != nil {
			pb.Enabled = *req.Enabled
			return s.repo.Update(pb)
		}
		return fmt.Errorf("内置剧本不可编辑")
	}

	pb.Name = req.Name
	pb.Description = req.Description
	if req.Enabled != nil {
		pb.Enabled = *req.Enabled
	}

	if err := s.repo.Update(pb); err != nil {
		return fmt.Errorf("更新剧本失败: %w", err)
	}

	// 替换步骤
	var steps []model.PlaybookStep
	for i, step := range req.Steps {
		timeout := step.Timeout
		if timeout <= 0 {
			timeout = 60
		}
		onError := step.OnError
		if onError == "" {
			onError = "stop"
		}
		steps = append(steps, model.PlaybookStep{
			PlaybookID: id,
			SortOrder:  i,
			Name:       step.Name,
			Script:     step.Script,
			Timeout:    timeout,
			OnError:    onError,
		})
	}

	return s.repo.ReplaceSteps(id, steps)
}

// Delete 删除剧本（内置剧本不可删除）
func (s *PlaybookService) Delete(id uint) error {
	pb, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("剧本不存在")
	}
	if pb.IsBuiltin {
		return fmt.Errorf("内置剧本不可删除")
	}
	return s.repo.Delete(id)
}

// playbookAssetJob 剧本执行资产任务
type playbookAssetJob struct {
	asset *model.Asset
	pwd   string
}

// Execute 执行剧本（支持多剧本合并执行，一次 SSH 连接跑完所有步骤）
func (s *PlaybookService) Execute(req dto.ExecutePlaybookRequest, userID uint) (*dto.PlaybookJobResponse, error) {
	// 加载并合并所有剧本的步骤
	var allSteps []model.PlaybookStep
	var names []string
	var ids []string
	for _, pbID := range req.PlaybookIDs {
		pb, err := s.repo.GetByID(pbID)
		if err != nil {
			return nil, fmt.Errorf("剧本 %d 不存在", pbID)
		}
		if !pb.Enabled {
			return nil, fmt.Errorf("剧本 %s 已禁用", pb.Name)
		}
		if len(pb.Steps) == 0 {
			return nil, fmt.Errorf("剧本 %s 没有步骤", pb.Name)
		}
		names = append(names, pb.Name)
		ids = append(ids, fmt.Sprintf("%d", pb.ID))
		allSteps = append(allSteps, pb.Steps...)
	}

	// 批量加载资产
	assetList, err := s.assetRepo.GetByIDs(req.AssetIDs)
	if err != nil {
		return nil, fmt.Errorf("加载资产失败: %w", err)
	}
	assetMap := make(map[uint]*model.Asset, len(assetList))
	for i := range assetList {
		assetMap[assetList[i].ID] = &assetList[i]
	}

	var assets []playbookAssetJob
	for _, id := range req.AssetIDs {
		asset, ok := assetMap[id]
		if !ok {
			continue
		}
		pwd := ""
		if asset.SshPassword != "" {
			pwd, _ = crypto.Decrypt(asset.SshPassword, s.masterKey)
		}
		assets = append(assets, playbookAssetJob{asset: asset, pwd: pwd})
	}

	if len(assets) == 0 {
		return nil, fmt.Errorf("没有有效的目标资产")
	}

	// 创建执行记录
	now := time.Now()
	exec := &model.PlaybookExecution{
		PlaybookID:   req.PlaybookIDs[0],
		PlaybookIDs:  strings.Join(ids, ","),
		PlaybookName: strings.Join(names, " + "),
		Status:       "running",
		TotalAssets:  len(assets),
		StartedAt:    now,
		CreatedBy:    userID,
	}
	if err := s.repo.CreateExecution(exec); err != nil {
		return nil, fmt.Errorf("创建执行记录失败: %w", err)
	}

	// 生成 job ID
	seq := playbookJobCounter.Add(1)
	jobID := fmt.Sprintf("playbook_%d", seq)

	// 启动异步执行
	go s.runPlaybook(jobID, exec, assets, allSteps)

	return &dto.PlaybookJobResponse{
		JobID:       jobID,
		ExecutionID: exec.ID,
		Total:       len(assets),
	}, nil
}

// runPlaybook 异步执行剧本
func (s *PlaybookService) runPlaybook(
	jobID string,
	exec *model.PlaybookExecution,
	assets []playbookAssetJob,
	steps []model.PlaybookStep,
) {
	timeout := time.Duration(s.configSvc.GetInt("ssh.timeout", 30)) * time.Second
	maxConcurrent := s.configSvc.GetInt("ssh.max_concurrent", 100)

	results := make([]dto.PlaybookExecuteResultItem, len(assets))
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i, job := range assets {
		wg.Add(1)
		go func(idx int, j playbookAssetJob) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := s.executeOnAsset(j.asset, j.pwd, steps, timeout)
			results[idx] = result

			// 保存单资产结果
			stepJSON, _ := json.Marshal(result.StepResults)
			s.repo.CreateExecutionResult(&model.PlaybookExecutionResult{
				ExecutionID: exec.ID,
				AssetID:     j.asset.ID,
				Hostname:    j.asset.Hostname,
				IP:          j.asset.IP,
				Status:      result.Status,
				StepResults: string(stepJSON),
			})

			// 实时推送
			s.hub.Broadcast(jobID, ws.Message{
				Type: "playbook_progress",
				Data: result,
			})
		}(i, job)
	}
	wg.Wait()

	// 统计
	resp := &dto.PlaybookExecuteResponse{
		Total:   len(results),
		Results: results,
	}
	for _, r := range results {
		if r.Status == "success" {
			resp.Success++
		} else {
			resp.Failed++
		}
	}

	// 更新执行记录
	now := time.Now()
	exec.FinishedAt = &now
	exec.Success = resp.Success
	exec.Failed = resp.Failed
	if resp.Failed == 0 {
		exec.Status = "success"
	} else if resp.Success > 0 {
		exec.Status = "partial"
	} else {
		exec.Status = "failed"
	}
	s.repo.UpdateExecution(exec)

	// 推送完成
	s.hub.Broadcast(jobID, ws.Message{
		Type: "playbook_done",
		Data: resp,
	})

	logger.Log.Infof("剧本执行 %s 完成: success=%d, failed=%d", jobID, resp.Success, resp.Failed)
}

// executeOnAsset 在单个资产上执行剧本所有步骤
func (s *PlaybookService) executeOnAsset(
	asset *model.Asset,
	password string,
	steps []model.PlaybookStep,
	connTimeout time.Duration,
) dto.PlaybookExecuteResultItem {
	result := dto.PlaybookExecuteResultItem{
		AssetID:  asset.ID,
		Hostname: asset.Hostname,
		IP:       asset.IP,
	}

	// 建立 SSH 连接
	client, err := sshutil.Dial(sshutil.ClientConfig{
		Host:     asset.IP,
		Port:     asset.Port,
		User:     asset.SshUser,
		Password: password,
		Timeout:  connTimeout,
	})
	if err != nil {
		result.Status = "failed"
		result.StepResults = []dto.StepExecResult{{
			Name:   "SSH 连接",
			Status: "failed",
			Error:  err.Error(),
		}}
		return result
	}
	defer client.Close()

	// 逐步执行
	totalSteps := len(steps)
	successSteps := 0
	stopped := false

	for _, step := range steps {
		if stopped {
			result.StepResults = append(result.StepResults, dto.StepExecResult{
				Name:   step.Name,
				Status: "skipped",
				Error:  "前置步骤失败，已跳过",
			})
			continue
		}

		stepResult := s.executeStep(client, step)
		result.StepResults = append(result.StepResults, stepResult)

		if stepResult.Status == "success" {
			successSteps++
		} else if step.OnError == "stop" {
			stopped = true
		}
	}

	if successSteps == totalSteps {
		result.Status = "success"
	} else if successSteps > 0 {
		result.Status = "partial"
	} else {
		result.Status = "failed"
	}

	return result
}

// executeStep 执行单个步骤
func (s *PlaybookService) executeStep(client *ssh.Client, step model.PlaybookStep) dto.StepExecResult {
	start := time.Now()
	timeout := time.Duration(step.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	r := sshutil.RunCommand(client, step.Script, timeout)
	latency := time.Since(start).Milliseconds()

	if r.Err != nil {
		return dto.StepExecResult{
			Name:    step.Name,
			Status:  "failed",
			Error:   r.Err.Error(),
			Output:  truncateOutput(r.Stdout, 2000),
			Latency: latency,
		}
	}
	if r.ExitCode != 0 {
		errMsg := r.Stderr
		if errMsg == "" {
			errMsg = fmt.Sprintf("exit code: %d", r.ExitCode)
		}
		return dto.StepExecResult{
			Name:    step.Name,
			Status:  "failed",
			Error:   truncateOutput(errMsg, 1000),
			Output:  truncateOutput(r.Stdout, 2000),
			Latency: latency,
		}
	}

	return dto.StepExecResult{
		Name:    step.Name,
		Status:  "success",
		Output:  truncateOutput(r.Stdout, 2000),
		Latency: latency,
	}
}

// ListExecutions 查询执行历史
func (s *PlaybookService) ListExecutions(page, pageSize int) ([]dto.PlaybookExecutionListItem, int64, error) {
	execs, total, err := s.repo.ListExecutions(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var items []dto.PlaybookExecutionListItem
	for _, e := range execs {
		finished := ""
		if e.FinishedAt != nil {
			finished = e.FinishedAt.Format("2006-01-02 15:04:05")
		}
		items = append(items, dto.PlaybookExecutionListItem{
			ID:           e.ID,
			PlaybookIDs:  e.PlaybookIDs,
			PlaybookName: e.PlaybookName,
			Status:       e.Status,
			TotalAssets:  e.TotalAssets,
			Success:      e.Success,
			Failed:       e.Failed,
			StartedAt:    e.StartedAt.Format("2006-01-02 15:04:05"),
			FinishedAt:   finished,
			CreatedBy:    e.CreatedBy,
		})
	}
	return items, total, nil
}

// GetExecutionResults 查询执行结果详情
func (s *PlaybookService) GetExecutionResults(executionID uint) ([]dto.PlaybookExecuteResultItem, error) {
	raw, err := s.repo.GetExecutionResults(executionID)
	if err != nil {
		return nil, err
	}

	var results []dto.PlaybookExecuteResultItem
	for _, r := range raw {
		item := dto.PlaybookExecuteResultItem{
			AssetID:  r.AssetID,
			Hostname: r.Hostname,
			IP:       r.IP,
			Status:   r.Status,
		}
		if r.StepResults != "" {
			json.Unmarshal([]byte(r.StepResults), &item.StepResults)
		}
		results = append(results, item)
	}
	return results, nil
}

// SeedBuiltinPlaybooks 初始化内置剧本
func (s *PlaybookService) SeedBuiltinPlaybooks() error {
	existing, _ := s.repo.List()
	builtinNames := make(map[string]bool)
	for _, pb := range existing {
		if pb.IsBuiltin {
			builtinNames[pb.Name] = true
		}
	}

	builtins := []model.Playbook{
		{
			Name:        "内核参数优化",
			Description: "优化 Linux 内核参数，包括网络、文件描述符、内存等常见调优项",
			Category:    "builtin",
			IsBuiltin:   true,
			Enabled:     true,
			Steps: []model.PlaybookStep{
				{SortOrder: 0, Name: "备份当前 sysctl 配置", Script: "cp /etc/sysctl.conf /etc/sysctl.conf.bak.$(date +%Y%m%d%H%M%S)", Timeout: 10, OnError: "continue"},
				{SortOrder: 1, Name: "优化网络参数", Script: `cat >> /etc/sysctl.conf << 'SYSCTL_EOF'
# --- MXCMDB 内核优化 ---
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 1200
net.ipv4.tcp_max_syn_backlog = 8192
net.ipv4.tcp_tw_reuse = 1
net.ipv4.ip_local_port_range = 1024 65535
net.core.somaxconn = 32768
net.core.netdev_max_backlog = 32768
SYSCTL_EOF`, Timeout: 30, OnError: "stop"},
				{SortOrder: 2, Name: "优化文件描述符限制", Script: `cat >> /etc/sysctl.conf << 'SYSCTL_EOF'
fs.file-max = 1048576
fs.nr_open = 1048576
SYSCTL_EOF
cat >> /etc/security/limits.conf << 'LIMITS_EOF'
# --- MXCMDB 内核优化 ---
* soft nofile 1048576
* hard nofile 1048576
* soft nproc 65535
* hard nproc 65535
LIMITS_EOF`, Timeout: 30, OnError: "stop"},
				{SortOrder: 3, Name: "优化内存参数", Script: `cat >> /etc/sysctl.conf << 'SYSCTL_EOF'
vm.swappiness = 10
vm.max_map_count = 262144
vm.overcommit_memory = 1
SYSCTL_EOF`, Timeout: 30, OnError: "stop"},
				{SortOrder: 4, Name: "应用 sysctl 配置", Script: "sysctl -p", Timeout: 30, OnError: "stop"},
				{SortOrder: 5, Name: "验证关键参数", Script: "sysctl net.ipv4.tcp_fin_timeout net.core.somaxconn fs.file-max vm.swappiness", Timeout: 10, OnError: "continue"},
			},
		},
		{
			Name:        "安全加固",
			Description: "Linux 基础安全加固：SSH 配置强化、密码策略、防火墙基础规则",
			Category:    "builtin",
			IsBuiltin:   true,
			Enabled:     true,
			Steps: []model.PlaybookStep{
				{SortOrder: 0, Name: "备份 SSH 配置", Script: "cp /etc/ssh/sshd_config /etc/ssh/sshd_config.bak.$(date +%Y%m%d%H%M%S)", Timeout: 10, OnError: "continue"},
				{SortOrder: 1, Name: "禁止 root SSH 登录", Script: `sed -i 's/^#*PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config`, Timeout: 15, OnError: "stop"},
				{SortOrder: 2, Name: "禁止密码空登录", Script: `sed -i 's/^#*PermitEmptyPasswords.*/PermitEmptyPasswords no/' /etc/ssh/sshd_config`, Timeout: 15, OnError: "continue"},
				{SortOrder: 3, Name: "设置 SSH 超时", Script: `sed -i 's/^#*ClientAliveInterval.*/ClientAliveInterval 300/' /etc/ssh/sshd_config
sed -i 's/^#*ClientAliveCountMax.*/ClientAliveCountMax 3/' /etc/ssh/sshd_config`, Timeout: 15, OnError: "continue"},
				{SortOrder: 4, Name: "设置密码策略", Script: `sed -i 's/^PASS_MAX_DAYS.*/PASS_MAX_DAYS   90/' /etc/login.defs 2>/dev/null
sed -i 's/^PASS_MIN_DAYS.*/PASS_MIN_DAYS   7/' /etc/login.defs 2>/dev/null
sed -i 's/^PASS_MIN_LEN.*/PASS_MIN_LEN    12/' /etc/login.defs 2>/dev/null`, Timeout: 15, OnError: "continue"},
				{SortOrder: 5, Name: "重启 SSH 服务", Script: "systemctl restart sshd", Timeout: 30, OnError: "stop"},
			},
		},
		{
			Name:        "系统信息采集",
			Description: "采集主机基础信息：系统、CPU、内存、磁盘、网络、运行服务",
			Category:    "builtin",
			IsBuiltin:   true,
			Enabled:     true,
			Steps: []model.PlaybookStep{
				{SortOrder: 0, Name: "系统基础信息", Script: `echo "=== 主机名 ===" && hostname
echo "=== 系统版本 ===" && cat /etc/os-release 2>/dev/null | grep -E "^(NAME|VERSION)=" || cat /etc/redhat-release 2>/dev/null
echo "=== 内核版本 ===" && uname -r
echo "=== 运行时间 ===" && uptime`, Timeout: 15, OnError: "continue"},
				{SortOrder: 1, Name: "CPU 与内存", Script: `echo "=== CPU ===" && lscpu | grep -E "^(Architecture|CPU\(s\)|Model name|Thread)"
echo "=== 内存 ===" && free -h`, Timeout: 15, OnError: "continue"},
				{SortOrder: 2, Name: "磁盘与挂载", Script: `echo "=== 磁盘使用 ===" && df -h
echo "=== 块设备 ===" && lsblk 2>/dev/null || fdisk -l 2>/dev/null | head -20`, Timeout: 15, OnError: "continue"},
				{SortOrder: 3, Name: "网络配置", Script: `echo "=== IP 地址 ===" && ip addr 2>/dev/null || ifconfig
echo "=== 默认路由 ===" && ip route show default 2>/dev/null
echo "=== DNS ===" && cat /etc/resolv.conf | grep nameserver`, Timeout: 15, OnError: "continue"},
				{SortOrder: 4, Name: "运行服务", Script: `echo "=== 运行中的服务 ===" && systemctl list-units --type=service --state=running --no-pager 2>/dev/null | head -30
echo "=== Docker ===" && docker version --format '{{.Server.Version}}' 2>/dev/null || echo "未安装 Docker"`, Timeout: 30, OnError: "continue"},
			},
		},
		{
			Name:        "时间同步配置",
			Description: "配置 NTP/Chrony 时间同步，确保服务器时钟准确",
			Category:    "builtin",
			IsBuiltin:   true,
			Enabled:     true,
			Steps: []model.PlaybookStep{
				{SortOrder: 0, Name: "检查现有时间同步", Script: `echo "=== 当前时间 ===" && date
echo "=== 时区 ===" && timedatectl 2>/dev/null | grep -E "(Time zone|NTP)"
echo "=== chrony/ntp 状态 ===" && (systemctl is-active chronyd 2>/dev/null || systemctl is-active ntpd 2>/dev/null || echo "未运行")`, Timeout: 15, OnError: "continue"},
				{SortOrder: 1, Name: "安装 chrony", Script: `if command -v apt-get &>/dev/null; then
  apt-get install -y chrony 2>/dev/null
elif command -v yum &>/dev/null; then
  yum install -y chrony 2>/dev/null
fi
echo "chrony 安装完成"`, Timeout: 120, OnError: "stop"},
				{SortOrder: 2, Name: "启动 chrony 服务", Script: "systemctl enable chronyd && systemctl restart chronyd", Timeout: 30, OnError: "stop"},
				{SortOrder: 3, Name: "验证同步状态", Script: "chronyc tracking 2>/dev/null && chronyc sources 2>/dev/null", Timeout: 15, OnError: "continue"},
			},
		},
	}

	for i := range builtins {
		if builtinNames[builtins[i].Name] {
			continue
		}
		if err := s.repo.Create(&builtins[i]); err != nil {
			logger.Log.Errorf("初始化内置剧本 %s 失败: %v", builtins[i].Name, err)
		}
	}
	return nil
}

// toResponse 转换为响应 DTO
func (s *PlaybookService) toResponse(pb *model.Playbook) dto.PlaybookResponse {
	var steps []dto.PlaybookStepConfig
	for _, step := range pb.Steps {
		steps = append(steps, dto.PlaybookStepConfig{
			Name:    step.Name,
			Script:  step.Script,
			Timeout: step.Timeout,
			OnError: step.OnError,
		})
	}
	return dto.PlaybookResponse{
		ID:          pb.ID,
		Name:        pb.Name,
		Description: pb.Description,
		Category:    pb.Category,
		IsBuiltin:   pb.IsBuiltin,
		Enabled:     pb.Enabled,
		Steps:       steps,
		CreatedBy:   pb.CreatedBy,
		CreatedAt:   pb.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   pb.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// truncateOutput 截断过长的输出
func truncateOutput(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...(truncated)"
}
