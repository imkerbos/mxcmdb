package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/pkg/crypto"
	"github.com/imkerbos/mxcmdb/internal/pkg/sshutil"
	"github.com/imkerbos/mxcmdb/internal/repository"
)

// AssetService 资产管理服务
type AssetService struct {
	repo            *repository.AssetRepository
	projectRepo     *repository.ProjectRepository
	sshKeyRepo      *repository.SSHKeyRepository
	probeRepo       *repository.ProbeResultRepository
	linuxUserRepo   *repository.LinuxUserRepository
	termSessionRepo *repository.TerminalSessionRepository
	masterKey       string
	configSvc       *SystemConfigService
}

// NewAssetService 创建 AssetService
func NewAssetService(repo *repository.AssetRepository, masterKey string, configSvc *SystemConfigService, projectRepo *repository.ProjectRepository, sshKeyRepo *repository.SSHKeyRepository) *AssetService {
	return &AssetService{
		repo:        repo,
		projectRepo: projectRepo,
		sshKeyRepo:  sshKeyRepo,
		masterKey:   masterKey,
		configSvc:   configSvc,
	}
}

// SetDetailRepos 注入详情页所需的额外 repository（避免改动现有构造函数签名）
func (s *AssetService) SetDetailRepos(probeRepo *repository.ProbeResultRepository, linuxUserRepo *repository.LinuxUserRepository, termSessionRepo *repository.TerminalSessionRepository) {
	s.probeRepo = probeRepo
	s.linuxUserRepo = linuxUserRepo
	s.termSessionRepo = termSessionRepo
}

// List 资产列表
func (s *AssetService) List(keyword, assetType, source, status, env, dept string, projectID uint, owner string, page, pageSize int) ([]dto.AssetResponse, int64, error) {
	assets, total, err := s.repo.List(keyword, assetType, source, status, env, dept, projectID, owner, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("查询资产列表失败: %w", err)
	}

	projectMap := s.buildProjectMap(assets)
	sshKeyMap := s.buildSSHKeyMap(assets)
	result := make([]dto.AssetResponse, len(assets))
	for i, a := range assets {
		result[i] = toAssetResponse(a, projectMap, sshKeyMap)
	}
	return result, total, nil
}

// GetByID 资产详情
func (s *AssetService) GetByID(id uint) (*dto.AssetResponse, error) {
	asset, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("资产不存在")
	}
	projectMap := s.buildProjectMap([]model.Asset{*asset})
	sshKeyMap := s.buildSSHKeyMap([]model.Asset{*asset})
	resp := toAssetResponse(*asset, projectMap, sshKeyMap)
	return &resp, nil
}

// GetDetail 资产详情（聚合探针、SSH Key、Linux 用户、终端会话）
func (s *AssetService) GetDetail(id uint) (*dto.AssetDetailResponse, error) {
	asset, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("资产不存在")
	}

	projectMap := s.buildProjectMap([]model.Asset{*asset})
	sshKeyMap := s.buildSSHKeyMap([]model.Asset{*asset})
	base := toAssetResponse(*asset, projectMap, sshKeyMap)

	detail := &dto.AssetDetailResponse{
		AssetResponse: base,
	}

	// 最新探针数据
	if s.probeRepo != nil {
		if probe, err := s.probeRepo.GetLatestByAssetID(id); err == nil {
			detail.Probe = &dto.AssetDetailProbe{
				Hostname:        probe.Hostname,
				CPU:             probe.CPU,
				Memory:          probe.Memory,
				Disk:            probe.Disk,
				Network:         probe.Network,
				OS:              probe.OS,
				Kernel:          probe.Kernel,
				DockerVersion:   probe.DockerVersion,
				RunningServices: probe.RunningServices,
				SSHUsers:        probe.SSHUsers,
				Uptime:          probe.Uptime,
				DNS:             probe.DNS,
				Gateway:         probe.Gateway,
				Manufacturer:    probe.Manufacturer,
				ProductModel:    probe.ProductModel,
				SerialNumber:    probe.SerialNumber,
				PublicIP:        probe.PublicIP,
				Processes:       probe.Processes,
				Listeners:       probe.Listeners,
				CollectedAt:     probe.CollectedAt.Format("2006-01-02 15:04:05"),
				Status:          probe.Status,
			}
		}
	}

	// SSH Key 绑定
	if s.sshKeyRepo != nil {
		if bindings, err := s.sshKeyRepo.ListBindingsByAssetID(id); err == nil {
			detail.SSHKeyBindings = make([]dto.AssetDetailSSHKeyBinding, len(bindings))
			for i, b := range bindings {
				item := dto.AssetDetailSSHKeyBinding{
					ID:       b.ID,
					SSHKeyID: b.SSHKeyID,
					Username: b.Username,
					Status:   b.Status,
				}
				if key, err := s.sshKeyRepo.GetByID(b.SSHKeyID); err == nil {
					item.KeyName = key.Name
				}
				if b.DeployedAt != nil {
					item.DeployedAt = b.DeployedAt.Format("2006-01-02 15:04:05")
				}
				detail.SSHKeyBindings[i] = item
			}
		}
	}
	if detail.SSHKeyBindings == nil {
		detail.SSHKeyBindings = []dto.AssetDetailSSHKeyBinding{}
	}

	// Linux 用户
	if s.linuxUserRepo != nil {
		if users, err := s.linuxUserRepo.ListByAssetID(id); err == nil {
			detail.LinuxUsers = make([]dto.AssetDetailLinuxUser, len(users))
			for i, u := range users {
				detail.LinuxUsers[i] = dto.AssetDetailLinuxUser{
					ID:       u.ID,
					Username: u.Username,
					UID:      u.UID,
					GID:      u.GID,
					Home:     u.Home,
					Shell:    u.Shell,
					Sudo:     u.Sudo,
					Status:   u.Status,
				}
			}
		}
	}
	if detail.LinuxUsers == nil {
		detail.LinuxUsers = []dto.AssetDetailLinuxUser{}
	}

	// 终端会话（最近 20 条）
	if s.termSessionRepo != nil {
		if sessions, err := s.termSessionRepo.ListByAssetID(id, 20); err == nil {
			detail.TerminalSessions = make([]dto.AssetDetailTerminalSession, len(sessions))
			for i, sess := range sessions {
				item := dto.AssetDetailTerminalSession{
					ID:        sess.ID,
					UserID:    sess.UserID,
					Username:  sess.Username,
					Status:    sess.Status,
					ClientIP:  sess.ClientIP,
					StartedAt: sess.StartedAt.Format("2006-01-02 15:04:05"),
				}
				if sess.FinishedAt != nil {
					item.FinishedAt = sess.FinishedAt.Format("2006-01-02 15:04:05")
				}
				detail.TerminalSessions[i] = item
			}
		}
	}
	if detail.TerminalSessions == nil {
		detail.TerminalSessions = []dto.AssetDetailTerminalSession{}
	}

	// 探针历史（最近 10 条）
	if s.probeRepo != nil {
		if history, err := s.probeRepo.ListByAssetID(id, 10); err == nil {
			detail.ProbeHistory = make([]dto.AssetDetailProbeHistory, len(history))
			for i, h := range history {
				detail.ProbeHistory[i] = dto.AssetDetailProbeHistory{
					ID:          h.ID,
					Status:      h.Status,
					OS:          h.OS,
					Kernel:      h.Kernel,
					CollectedAt: h.CollectedAt.Format("2006-01-02 15:04:05"),
				}
			}
		}
	}
	if detail.ProbeHistory == nil {
		detail.ProbeHistory = []dto.AssetDetailProbeHistory{}
	}

	return detail, nil
}

// Create 创建资产
func (s *AssetService) Create(req dto.CreateAssetRequest) error {
	// 检查 IP 是否已存在
	if existing, _ := s.repo.GetByIP(req.IP); existing != nil {
		return fmt.Errorf("IP %s 已存在", req.IP)
	}

	asset := &model.Asset{
		Hostname:      req.Hostname,
		IP:            req.IP,
		Port:          req.Port,
		OS:            req.OS,
		Type:          req.Type,
		Source:        req.Source,
		Status:        req.Status,
		Spec:          req.Spec,
		Department:    req.Department,
		ProjectID:     req.ProjectID,
		Owner:         req.Owner,
		Environment:   req.Environment,
		BusinessGroup: req.BusinessGroup,
		SshUser:       req.SshUser,
		SshKeyID:      req.SshKeyID,
	}

	if asset.Port == 0 {
		asset.Port = 22
	}
	if asset.Source == "" {
		asset.Source = "manual"
	}
	if asset.Status == "" {
		asset.Status = "unknown"
	}
	if asset.SshUser == "" {
		asset.SshUser = "root"
	}

	// 加密 SSH 密码
	if req.SshPassword != "" {
		encrypted, err := crypto.Encrypt(req.SshPassword, s.masterKey)
		if err != nil {
			return fmt.Errorf("加密 SSH 密码失败: %w", err)
		}
		asset.SshPassword = encrypted
	}

	// 标签
	for _, t := range req.Tags {
		asset.Tags = append(asset.Tags, model.AssetTag{Key: t.Key, Value: t.Value})
	}

	return s.repo.Create(asset)
}

// Update 更新资产
func (s *AssetService) Update(id uint, req dto.UpdateAssetRequest) error {
	asset, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("资产不存在")
	}

	if req.Hostname != "" {
		asset.Hostname = req.Hostname
	}
	if req.Port > 0 {
		asset.Port = req.Port
	}
	if req.OS != "" {
		asset.OS = req.OS
	}
	if req.Type != "" {
		asset.Type = req.Type
	}
	if req.Status != "" {
		asset.Status = req.Status
	}
	if req.Spec != "" {
		asset.Spec = req.Spec
	}
	if req.Department != "" {
		asset.Department = req.Department
	}
	if req.ProjectID != nil {
		asset.ProjectID = req.ProjectID
	}
	if req.Owner != "" {
		asset.Owner = req.Owner
	}
	if req.Environment != "" {
		asset.Environment = req.Environment
	}
	if req.BusinessGroup != "" {
		asset.BusinessGroup = req.BusinessGroup
	}
	if req.SshUser != "" {
		asset.SshUser = req.SshUser
	}
	if req.SshPassword != "" {
		encrypted, err := crypto.Encrypt(req.SshPassword, s.masterKey)
		if err != nil {
			return fmt.Errorf("加密 SSH 密码失败: %w", err)
		}
		asset.SshPassword = encrypted
	}
	if req.SshKeyID != nil {
		asset.SshKeyID = req.SshKeyID
	}
	if req.ClearSshKey {
		asset.SshKeyID = nil
	}

	if err := s.repo.Update(asset); err != nil {
		return err
	}

	// 更新标签
	if req.Tags != nil {
		if err := s.repo.DeleteTagsByAssetID(id); err != nil {
			return fmt.Errorf("清除旧标签失败: %w", err)
		}
		tags := make([]model.AssetTag, len(req.Tags))
		for i, t := range req.Tags {
			tags[i] = model.AssetTag{AssetID: id, Key: t.Key, Value: t.Value}
		}
		if err := s.repo.CreateTags(tags); err != nil {
			return fmt.Errorf("创建标签失败: %w", err)
		}
	}

	return nil
}

// Delete 删除资产
func (s *AssetService) Delete(id uint) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return fmt.Errorf("资产不存在")
	}
	return s.repo.Delete(id)
}

// GetStats 获取资产统计
func (s *AssetService) GetStats() (*dto.AssetStatsResponse, error) {
	totalCloud, _ := s.repo.CountBySource("cloud_sync")
	totalIDC, _ := s.repo.CountBySource("manual")
	totalOnline, _ := s.repo.CountByStatus("online")
	totalOffline, _ := s.repo.CountByStatus("offline")
	totalProbed, _ := s.repo.CountProbed()

	return &dto.AssetStatsResponse{
		TotalCloud:   totalCloud,
		TotalIDC:     totalIDC,
		TotalOnline:  totalOnline,
		TotalOffline: totalOffline,
		TotalProbed:  totalProbed,
	}, nil
}

// ListAll 获取所有资产（用于选择器）
func (s *AssetService) ListAll() ([]model.Asset, error) {
	return s.repo.ListAll()
}

// BatchImport 批量导入资产
func (s *AssetService) BatchImport(req dto.BatchImportAssetRequest) (*dto.BatchImportAssetResponse, error) {
	resp := &dto.BatchImportAssetResponse{Total: len(req.Assets)}

	for i, item := range req.Assets {
		if err := s.Create(item); err != nil {
			resp.Failed++
			resp.Errors = append(resp.Errors, dto.BatchImportErrorItem{
				Index:   i,
				IP:      item.IP,
				Message: err.Error(),
			})
		} else {
			resp.Success++
		}
	}

	return resp, nil
}

// TestConnection 批量测试 SSH 连接
func (s *AssetService) TestConnection(assetIDs []uint) (*dto.BatchTestConnectionResponse, error) {
	timeout := 10 * time.Second
	if s.configSvc != nil {
		timeout = time.Duration(s.configSvc.GetInt("ssh.timeout", 10)) * time.Second
	}

	type connJob struct {
		asset *model.Asset
		index int
	}

	var jobs []connJob
	for i, id := range assetIDs {
		asset, err := s.repo.GetByID(id)
		if err != nil {
			continue
		}
		jobs = append(jobs, connJob{asset: asset, index: i})
	}

	// 预加载密钥（避免并发时重复查询）
	keyCache := make(map[uint]string)
	for _, j := range jobs {
		if j.asset.SshKeyID != nil && *j.asset.SshKeyID > 0 {
			if _, ok := keyCache[*j.asset.SshKeyID]; !ok {
				if key, err := s.sshKeyRepo.GetByID(*j.asset.SshKeyID); err == nil {
					privKey, err := crypto.Decrypt(key.PrivateKey, s.masterKey)
					if err == nil {
						keyCache[*j.asset.SshKeyID] = privKey
					}
				}
			}
		}
	}

	results := make([]dto.TestConnectionResult, len(jobs))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 20)

	for i, job := range jobs {
		wg.Add(1)
		go func(idx int, j connJob) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			password := ""
			if j.asset.SshPassword != "" {
				password, _ = crypto.Decrypt(j.asset.SshPassword, s.masterKey)
			}

			privateKey := ""
			if j.asset.SshKeyID != nil && *j.asset.SshKeyID > 0 {
				privateKey = keyCache[*j.asset.SshKeyID]
			}

			// 确定认证方式描述
			authMethod := "none"
			if privateKey != "" && password != "" {
				authMethod = "password+key"
			} else if privateKey != "" {
				authMethod = "key"
			} else if password != "" {
				authMethod = "password"
			}

			start := time.Now()
			client, err := sshutil.Dial(sshutil.ClientConfig{
				Host:       j.asset.IP,
				Port:       j.asset.Port,
				User:       j.asset.SshUser,
				Password:   password,
				PrivateKey: privateKey,
				Timeout:    timeout,
			})
			latency := time.Since(start).Milliseconds()

			r := dto.TestConnectionResult{
				AssetID:    j.asset.ID,
				Hostname:   j.asset.Hostname,
				IP:         j.asset.IP,
				Latency:    latency,
				AuthMethod: authMethod,
			}
			if err != nil {
				r.Status = "failed"
				r.Error = err.Error()
			} else {
				r.Status = "success"
				client.Close()
			}
			results[idx] = r
		}(i, job)
	}
	wg.Wait()

	// 根据连通性结果更新资产状态（仅 IDC 资产）
	for _, r := range results {
		newStatus := "offline"
		if r.Status == "success" {
			newStatus = "online"
		}
		if asset, err := s.repo.GetByID(r.AssetID); err == nil && asset.Source == "manual" {
			asset.Status = newStatus
			_ = s.repo.Update(asset)
		}
	}

	return &dto.BatchTestConnectionResponse{Results: results}, nil
}

// BatchDelete 批量删除资产
func (s *AssetService) BatchDelete(ids []uint) (*dto.BatchDeleteAssetResponse, error) {
	resp := &dto.BatchDeleteAssetResponse{Total: len(ids)}
	for _, id := range ids {
		if err := s.Delete(id); err != nil {
			resp.Failed++
		} else {
			resp.Success++
		}
	}
	return resp, nil
}

// buildProjectMap 根据资产列表批量查询项目名称
func (s *AssetService) buildProjectMap(assets []model.Asset) map[uint]string {
	m := make(map[uint]string)
	if s.projectRepo == nil {
		return m
	}
	for _, a := range assets {
		if a.ProjectID != nil && *a.ProjectID > 0 {
			m[*a.ProjectID] = ""
		}
	}
	for id := range m {
		if p, err := s.projectRepo.GetByID(id); err == nil {
			m[id] = p.Name
		}
	}
	return m
}

// buildSSHKeyMap 根据资产列表批量查询 SSH Key 名称
func (s *AssetService) buildSSHKeyMap(assets []model.Asset) map[uint]string {
	m := make(map[uint]string)
	if s.sshKeyRepo == nil {
		return m
	}
	for _, a := range assets {
		if a.SshKeyID != nil && *a.SshKeyID > 0 {
			m[*a.SshKeyID] = ""
		}
	}
	for id := range m {
		if key, err := s.sshKeyRepo.GetByID(id); err == nil {
			m[id] = key.Name
		}
	}
	return m
}

// toAssetResponse 转换为响应 DTO
func toAssetResponse(a model.Asset, projectMap map[uint]string, sshKeyMap map[uint]string) dto.AssetResponse {
	resp := dto.AssetResponse{
		ID:             a.ID,
		Hostname:       a.Hostname,
		IP:             a.IP,
		Port:           a.Port,
		OS:             a.OS,
		Type:           a.Type,
		Source:         a.Source,
		CloudAccountID: a.CloudAccountID,
		InstanceID:     a.InstanceID,
		Region:         a.Region,
		Zone:           a.Zone,
		Status:         a.Status,
		Spec:           a.Spec,
		Department:     a.Department,
		ProjectID:      a.ProjectID,
		Owner:          a.Owner,
		Environment:    a.Environment,
		BusinessGroup:  a.BusinessGroup,
		SshUser:        a.SshUser,
		SshKeyID:       a.SshKeyID,
		HasPassword:    a.SshPassword != "",
		CreatedAt:      a.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      a.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if a.ProjectID != nil {
		resp.ProjectName = projectMap[*a.ProjectID]
	}
	if a.SshKeyID != nil {
		resp.SshKeyName = sshKeyMap[*a.SshKeyID]
	}
	if a.ProbeLastAt != nil {
		resp.ProbeLastAt = a.ProbeLastAt.Format("2006-01-02 15:04:05")
	}
	resp.Tags = make([]dto.AssetTagItem, len(a.Tags))
	for i, t := range a.Tags {
		resp.Tags[i] = dto.AssetTagItem{Key: t.Key, Value: t.Value}
	}
	return resp
}
