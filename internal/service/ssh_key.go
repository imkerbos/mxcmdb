package service

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/pkg/crypto"
	"github.com/imkerbos/mxcmdb/internal/pkg/logger"
	"github.com/imkerbos/mxcmdb/internal/pkg/sshutil"
	"github.com/imkerbos/mxcmdb/internal/repository"
	"golang.org/x/crypto/ssh"
)

// SSHKeyService SSH Key 管理服务
type SSHKeyService struct {
	repo      *repository.SSHKeyRepository
	assetRepo *repository.AssetRepository
	masterKey string
	configSvc *SystemConfigService
	notifySvc *NotificationService
}

// SetNotificationService 注入通知服务
func (s *SSHKeyService) SetNotificationService(svc *NotificationService) {
	s.notifySvc = svc
}

// NewSSHKeyService 创建 SSHKeyService
func NewSSHKeyService(repo *repository.SSHKeyRepository, assetRepo *repository.AssetRepository, masterKey string, configSvc *SystemConfigService) *SSHKeyService {
	return &SSHKeyService{
		repo:      repo,
		assetRepo: assetRepo,
		masterKey: masterKey,
		configSvc: configSvc,
	}
}

// List SSH Key 列表
func (s *SSHKeyService) List(keyword string, page, pageSize int) ([]dto.SSHKeyResponse, int64, error) {
	keys, total, err := s.repo.List(keyword, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.SSHKeyResponse, len(keys))
	for i, k := range keys {
		result[i] = dto.SSHKeyResponse{
			ID:          k.ID,
			Name:        k.Name,
			PublicKey:   k.PublicKey,
			Fingerprint: k.Fingerprint,
			KeyType:     k.KeyType,
			Comment:     k.Comment,
			CreatedAt:   k.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return result, total, nil
}

// Create 创建/生成 SSH Key
func (s *SSHKeyService) Create(req dto.CreateSSHKeyRequest, createdBy uint) (*dto.SSHKeyResponse, error) {
	var publicKey, privateKey string
	var fingerprint, keyType string

	if req.PublicKey != "" && req.PrivateKey != "" {
		// 导入模式
		publicKey = req.PublicKey
		privateKey = req.PrivateKey

		pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
		if err != nil {
			return nil, fmt.Errorf("无效的公钥格式: %w", err)
		}
		fingerprint = ssh.FingerprintSHA256(pk)
		keyType = pk.Type()
	} else {
		// 生成模式：根据 key_type 生成不同类型的密钥
		switch req.KeyType {
		case "rsa", "ssh-rsa":
			rsaKey, err := rsa.GenerateKey(rand.Reader, 4096)
			if err != nil {
				return nil, fmt.Errorf("生成 RSA 密钥失败: %w", err)
			}
			sshPubKey, err := ssh.NewPublicKey(&rsaKey.PublicKey)
			if err != nil {
				return nil, fmt.Errorf("转换公钥失败: %w", err)
			}
			publicKey = string(ssh.MarshalAuthorizedKey(sshPubKey))
			fingerprint = ssh.FingerprintSHA256(sshPubKey)
			keyType = "ssh-rsa"
			privKeyBytes, err := ssh.MarshalPrivateKey(rsaKey, req.Comment)
			if err != nil {
				return nil, fmt.Errorf("序列化私钥失败: %w", err)
			}
			privateKey = string(pem.EncodeToMemory(privKeyBytes))
		default:
			// 默认 ED25519
			pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				return nil, fmt.Errorf("生成密钥失败: %w", err)
			}
			sshPubKey, err := ssh.NewPublicKey(pubKey)
			if err != nil {
				return nil, fmt.Errorf("转换公钥失败: %w", err)
			}
			publicKey = string(ssh.MarshalAuthorizedKey(sshPubKey))
			fingerprint = ssh.FingerprintSHA256(sshPubKey)
			keyType = "ssh-ed25519"
			privKeyBytes, err := ssh.MarshalPrivateKey(privKey, req.Comment)
			if err != nil {
				return nil, fmt.Errorf("序列化私钥失败: %w", err)
			}
			privateKey = string(pem.EncodeToMemory(privKeyBytes))
		}
	}

	// 加密私钥
	encryptedPrivKey, err := crypto.Encrypt(privateKey, s.masterKey)
	if err != nil {
		return nil, fmt.Errorf("加密私钥失败: %w", err)
	}

	key := &model.SSHKey{
		Name:        req.Name,
		PublicKey:   publicKey,
		PrivateKey:  encryptedPrivKey,
		Fingerprint: fingerprint,
		KeyType:     keyType,
		Comment:     req.Comment,
		CreatedBy:   createdBy,
	}

	if err := s.repo.Create(key); err != nil {
		return nil, fmt.Errorf("创建 SSH Key 失败: %w", err)
	}

	return &dto.SSHKeyResponse{
		ID:          key.ID,
		Name:        key.Name,
		PublicKey:   key.PublicKey,
		Fingerprint: key.Fingerprint,
		KeyType:     key.KeyType,
		Comment:     key.Comment,
		CreatedAt:   key.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// Delete 删除 SSH Key
func (s *SSHKeyService) Delete(id uint) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return fmt.Errorf("SSH Key 不存在")
	}
	_ = s.repo.DeleteBindingsByKeyID(id)
	return s.repo.Delete(id)
}

type deployTarget struct {
	task  sshutil.Task
	asset *model.Asset
}

// prepareDeployTargets 构建部署目标列表并创建绑定记录
func (s *SSHKeyService) prepareDeployTargets(keyID uint, req dto.DeploySSHKeyRequest) ([]deployTarget, []sshutil.Task) {
	// 批量加载资产
	assetList, _ := s.assetRepo.GetByIDs(req.AssetIDs)
	assetMap := make(map[uint]*model.Asset, len(assetList))
	for i := range assetList {
		assetMap[assetList[i].ID] = &assetList[i]
	}

	var targets []deployTarget
	var tasks []sshutil.Task
	for _, assetID := range req.AssetIDs {
		asset, ok := assetMap[assetID]
		if !ok {
			continue
		}
		password := ""
		if asset.SshPassword != "" {
			password, _ = crypto.Decrypt(asset.SshPassword, s.masterKey)
		}
		task := sshutil.Task{
			Host:     asset.IP,
			Port:     asset.Port,
			User:     asset.SshUser,
			Password: password,
		}
		targets = append(targets, deployTarget{task: task, asset: asset})
		tasks = append(tasks, task)

		// 创建绑定记录
		binding := &model.SSHKeyBinding{
			SSHKeyID: keyID,
			AssetID:  assetID,
			Username: req.Username,
			Status:   "pending",
		}
		_ = s.repo.CreateBinding(binding)
	}
	return targets, tasks
}

// deployExec 同步执行 SSH Key 部署，更新绑定状态并记录日志
func (s *SSHKeyService) deployExec(key *model.SSHKey, targets []deployTarget, tasks []sshutil.Task, req dto.DeploySSHKeyRequest, operatorID uint) {
	timeout := time.Duration(s.configSvc.GetInt("ssh.timeout", 30)) * time.Second
	maxConcurrent := s.configSvc.GetInt("ssh.max_concurrent", 100)
	pool := sshutil.NewWorkerPool(maxConcurrent)

	cmd := fmt.Sprintf(`mkdir -p ~%s/.ssh && chmod 700 ~%s/.ssh && echo '%s' >> ~%s/.ssh/authorized_keys && chmod 600 ~%s/.ssh/authorized_keys`,
		req.Username, req.Username, key.PublicKey, req.Username, req.Username)

	results := pool.Execute(tasks, func(task sshutil.Task) sshutil.ExecResult {
		client, err := sshutil.Dial(sshutil.ClientConfig{
			Host: task.Host, Port: task.Port, User: task.User,
			Password: task.Password, Timeout: timeout,
		})
		if err != nil {
			return sshutil.ExecResult{Err: err}
		}
		defer client.Close()
		return sshutil.RunCommand(client, cmd, timeout)
	})

	// 更新绑定状态 & 记录部署日志
	bindings, _ := s.repo.ListBindingsByKeyID(key.ID)
	var logs []model.SSHKeyDeployLog
	for i, r := range results {
		if i < len(bindings) {
			if r.Result.Err == nil && r.Result.ExitCode == 0 {
				now := time.Now()
				_ = s.repo.UpdateBindingStatus(bindings[i].ID, "deployed", map[string]interface{}{
					"deployed_at": now,
				})
			} else {
				_ = s.repo.UpdateBindingStatus(bindings[i].ID, "failed", nil)
			}
		}
		log := model.SSHKeyDeployLog{
			SSHKeyID:   key.ID,
			SSHKeyName: key.Name,
			AssetID:    targets[i].asset.ID,
			Hostname:   targets[i].asset.Hostname,
			IP:         targets[i].asset.IP,
			Username:   req.Username,
			Action:     "deploy",
			OperatorID: operatorID,
		}
		if r.Result.Err == nil && r.Result.ExitCode == 0 {
			log.Status = "success"
		} else {
			log.Status = "failed"
			if r.Result.Err != nil {
				log.Error = r.Result.Err.Error()
			} else if r.Result.Stderr != "" {
				log.Error = r.Result.Stderr
			}
		}
		logs = append(logs, log)
	}
	_ = s.repo.CreateDeployLogs(logs)
	logger.Log.Infof("SSH Key %d 部署完成，%d 台主机", key.ID, len(results))

	// 发送通知
	if s.notifySvc != nil {
		successCount := 0
		for _, r := range results {
			if r.Result.Err == nil && r.Result.ExitCode == 0 {
				successCount++
			}
		}
		failCount := len(results) - successCount
		nType := "success"
		if failCount > 0 {
			nType = "warning"
		}
		title := fmt.Sprintf("SSH Key「%s」部署完成", key.Name)
		content := fmt.Sprintf("成功 %d 台，失败 %d 台", successCount, failCount)
		s.notifySvc.NotifyAllOperators(title, content, nType, "ssh_key")
	}
}

// Deploy 部署 SSH Key 到资产（异步）
func (s *SSHKeyService) Deploy(keyID uint, req dto.DeploySSHKeyRequest, operatorID uint) error {
	key, err := s.repo.GetByID(keyID)
	if err != nil {
		return fmt.Errorf("SSH Key 不存在")
	}

	targets, tasks := s.prepareDeployTargets(keyID, req)

	go func() {
		s.deployExec(key, targets, tasks, req, operatorID)
	}()

	return nil
}

// Revoke 从资产撤销（回收）SSH Key
func (s *SSHKeyService) Revoke(keyID uint, req dto.RevokeSSHKeyRequest, operatorID uint) (*dto.RevokeResponse, error) {
	key, err := s.repo.GetByID(keyID)
	if err != nil {
		return nil, fmt.Errorf("SSH Key 不存在")
	}

	timeout := time.Duration(s.configSvc.GetInt("ssh.timeout", 30)) * time.Second
	maxConcurrent := s.configSvc.GetInt("ssh.max_concurrent", 100)
	pool := sshutil.NewWorkerPool(maxConcurrent)

	type revokeTarget struct {
		task    sshutil.Task
		asset   *model.Asset
		binding *model.SSHKeyBinding
	}

	var targets []revokeTarget
	var tasks []sshutil.Task

	// 批量加载资产
	revokeAssetList, _ := s.assetRepo.GetByIDs(req.AssetIDs)
	revokeAssetMap := make(map[uint]*model.Asset, len(revokeAssetList))
	for i := range revokeAssetList {
		revokeAssetMap[revokeAssetList[i].ID] = &revokeAssetList[i]
	}

	for _, assetID := range req.AssetIDs {
		asset, ok := revokeAssetMap[assetID]
		if !ok {
			continue
		}
		password := ""
		if asset.SshPassword != "" {
			password, _ = crypto.Decrypt(asset.SshPassword, s.masterKey)
		}
		task := sshutil.Task{
			Host:     asset.IP,
			Port:     asset.Port,
			User:     asset.SshUser,
			Password: password,
		}
		binding, _ := s.repo.GetBindingByKeyAssetUser(keyID, assetID, req.Username)
		targets = append(targets, revokeTarget{task: task, asset: asset, binding: binding})
		tasks = append(tasks, task)
	}

	// 构造移除公钥的命令：从 authorized_keys 中精确删除该公钥行
	// 取公钥核心内容（去掉尾部换行和注释部分仅匹配 key data）
	pubKeyContent := strings.TrimSpace(key.PublicKey)
	// 使用 grep -v 精确移除包含该公钥的行
	cmd := fmt.Sprintf(`sudo grep -v '%s' ~%s/.ssh/authorized_keys > ~%s/.ssh/authorized_keys.tmp && sudo mv ~%s/.ssh/authorized_keys.tmp ~%s/.ssh/authorized_keys && sudo chmod 600 ~%s/.ssh/authorized_keys`,
		escapeForShell(pubKeyContent), req.Username, req.Username, req.Username, req.Username, req.Username)

	results := pool.Execute(tasks, func(task sshutil.Task) sshutil.ExecResult {
		client, err := sshutil.Dial(sshutil.ClientConfig{
			Host: task.Host, Port: task.Port, User: task.User,
			Password: task.Password, Timeout: timeout,
		})
		if err != nil {
			return sshutil.ExecResult{Err: err}
		}
		defer client.Close()
		return sshutil.RunCommand(client, cmd, timeout)
	})

	resp := &dto.RevokeResponse{
		Total:   len(results),
		Results: make([]dto.RevokeResultItem, len(results)),
	}

	var logs []model.SSHKeyDeployLog
	for i, r := range results {
		item := dto.RevokeResultItem{
			AssetID:  targets[i].asset.ID,
			Hostname: targets[i].asset.Hostname,
			IP:       targets[i].asset.IP,
		}
		log := model.SSHKeyDeployLog{
			SSHKeyID:   keyID,
			SSHKeyName: key.Name,
			AssetID:    targets[i].asset.ID,
			Hostname:   targets[i].asset.Hostname,
			IP:         targets[i].asset.IP,
			Username:   req.Username,
			Action:     "revoke",
			OperatorID: operatorID,
		}
		if r.Result.Err == nil && r.Result.ExitCode == 0 {
			item.Status = "success"
			log.Status = "success"
			resp.Success++
			// 更新绑定状态
			if targets[i].binding != nil {
				_ = s.repo.UpdateBindingStatus(targets[i].binding.ID, "revoked", nil)
			}
		} else {
			item.Status = "failed"
			log.Status = "failed"
			resp.Failed++
			errMsg := ""
			if r.Result.Err != nil {
				errMsg = r.Result.Err.Error()
			} else if r.Result.Stderr != "" {
				errMsg = r.Result.Stderr
			}
			item.Error = errMsg
			log.Error = errMsg
		}
		resp.Results[i] = item
		logs = append(logs, log)
	}
	_ = s.repo.CreateDeployLogs(logs)

	logger.Log.Infof("SSH Key %d 撤销完成: 成功 %d, 失败 %d", keyID, resp.Success, resp.Failed)
	return resp, nil
}

// Rotate SSH Key 轮换：选择/生成新密钥 → 部署到所有已绑定资产 → 撤销旧密钥
func (s *SSHKeyService) Rotate(keyID uint, req dto.RotateSSHKeyRequest, createdBy uint) (*dto.RotateResponse, error) {
	oldKey, err := s.repo.GetByID(keyID)
	if err != nil {
		return nil, fmt.Errorf("SSH Key 不存在")
	}

	// 查询旧密钥的所有已部署绑定
	bindings, err := s.repo.ListDeployedBindingsByKeyID(keyID)
	if err != nil {
		return nil, fmt.Errorf("查询绑定关系失败: %w", err)
	}
	if len(bindings) == 0 {
		return nil, fmt.Errorf("该密钥没有已部署的绑定，无需轮换")
	}

	var newKey *dto.SSHKeyResponse

	if req.NewKeyID != nil {
		// 使用已有密钥
		existingKey, err := s.repo.GetByID(*req.NewKeyID)
		if err != nil {
			return nil, fmt.Errorf("选择的替换密钥不存在")
		}
		if existingKey.ID == keyID {
			return nil, fmt.Errorf("替换密钥不能与当前密钥相同")
		}
		newKey = &dto.SSHKeyResponse{
			ID:          existingKey.ID,
			Name:        existingKey.Name,
			PublicKey:   existingKey.PublicKey,
			Fingerprint: existingKey.Fingerprint,
			KeyType:     existingKey.KeyType,
			Comment:     existingKey.Comment,
			CreatedAt:   existingKey.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	} else {
		// 自动生成新密钥
		newName := req.NewName
		if newName == "" {
			newName = fmt.Sprintf("%s-rotated-%s", oldKey.Name, time.Now().Format("20060102"))
		}
		newKey, err = s.Create(dto.CreateSSHKeyRequest{
			Name:    newName,
			Comment: req.NewComment,
		}, createdBy)
		if err != nil {
			return nil, fmt.Errorf("生成新密钥失败: %w", err)
		}
	}

	// 获取新密钥的完整记录（deployExec 需要 model.SSHKey）
	newKeyModel, err := s.repo.GetByID(newKey.ID)
	if err != nil {
		return nil, fmt.Errorf("新密钥查询失败: %w", err)
	}

	resp := &dto.RotateResponse{
		OldKeyID: keyID,
		NewKey:   newKey,
	}

	// 按 username 分组绑定
	usernameAssets := make(map[string][]uint)
	for _, b := range bindings {
		usernameAssets[b.Username] = append(usernameAssets[b.Username], b.AssetID)
	}

	// 预先创建绑定并构建部署目标
	type deployGroup struct {
		req     dto.DeploySSHKeyRequest
		targets []deployTarget
		tasks   []sshutil.Task
	}
	var groups []deployGroup
	for username, assetIDs := range usernameAssets {
		deployReq := dto.DeploySSHKeyRequest{
			AssetIDs: assetIDs,
			Username: username,
		}
		targets, tasks := s.prepareDeployTargets(newKey.ID, deployReq)
		groups = append(groups, deployGroup{req: deployReq, targets: targets, tasks: tasks})
	}

	// 单个 goroutine 协调：先同步部署新密钥，再同步撤销旧密钥
	go func() {
		// 1. 同步部署新密钥到所有资产
		for _, g := range groups {
			s.deployExec(newKeyModel, g.targets, g.tasks, g.req, createdBy)
		}
		logger.Log.Infof("SSH Key 轮换: 新密钥 %d 部署完成，开始撤销旧密钥 %d", newKey.ID, keyID)

		// 2. 新密钥部署完成后，同步撤销旧密钥
		for username, assetIDs := range usernameAssets {
			_, err := s.Revoke(keyID, dto.RevokeSSHKeyRequest{
				AssetIDs: assetIDs,
				Username: username,
			}, createdBy)
			if err != nil {
				logger.Log.Errorf("轮换密钥 %d 撤销失败: %v", keyID, err)
			}
		}
		logger.Log.Infof("SSH Key %d 轮换完成，新密钥 ID: %d", keyID, newKey.ID)

		// 发送轮换完成通知
		if s.notifySvc != nil {
			title := fmt.Sprintf("SSH Key「%s」轮换完成", oldKey.Name)
			content := fmt.Sprintf("旧密钥已撤销，新密钥「%s」已部署到 %d 台资产", newKeyModel.Name, len(bindings))
			s.notifySvc.NotifyAllOperators(title, content, "success", "ssh_key")
		}
	}()

	return resp, nil
}

// DepartureCleanup 离职清理：撤销指定 SSH Key 在所有/指定资产上的部署
func (s *SSHKeyService) DepartureCleanup(req dto.DepartureCleanupRequest, operatorID uint) (*dto.DepartureCleanupResponse, error) {
	// 指定资产过滤
	targetAssetIDs := make(map[uint]bool)
	if len(req.AssetIDs) > 0 {
		for _, id := range req.AssetIDs {
			targetAssetIDs[id] = true
		}
	}

	resp := &dto.DepartureCleanupResponse{}
	totalAssets := 0

	for _, keyID := range req.KeyIDs {
		// 查询该密钥的所有已部署绑定
		bindings, err := s.repo.ListDeployedBindingsByKeyID(keyID)
		if err != nil {
			logger.Log.Errorf("离职清理查询密钥 %d 绑定失败: %v", keyID, err)
			continue
		}
		if len(bindings) == 0 {
			continue
		}

		// 按 username 分组（同一个 key 可能部署到不同 username）
		usernameAssets := make(map[string][]uint)
		for _, b := range bindings {
			if len(targetAssetIDs) > 0 && !targetAssetIDs[b.AssetID] {
				continue
			}
			usernameAssets[b.Username] = append(usernameAssets[b.Username], b.AssetID)
		}

		for username, assetIDs := range usernameAssets {
			totalAssets += len(assetIDs)
			revokeResp, err := s.Revoke(keyID, dto.RevokeSSHKeyRequest{
				AssetIDs: assetIDs,
				Username: username,
			}, operatorID)
			if err != nil {
				logger.Log.Errorf("离职清理密钥 %d 用户 %s 撤销失败: %v", keyID, username, err)
				continue
			}
			resp.RevokeResults = append(resp.RevokeResults, *revokeResp)
		}
	}

	resp.TotalKeys = len(req.KeyIDs)
	resp.TotalAssets = totalAssets

	if totalAssets == 0 {
		return nil, fmt.Errorf("选中的密钥没有已部署的绑定，无需清理")
	}

	logger.Log.Infof("离职清理完成: 密钥=%d, 资产=%d", resp.TotalKeys, resp.TotalAssets)
	return resp, nil
}

// escapeForShell 转义 shell 特殊字符用于单引号内使用
func escapeForShell(s string) string {
	return strings.ReplaceAll(s, "'", "'\"'\"'")
}

// Download 下载 SSH Key（解密私钥）
func (s *SSHKeyService) Download(keyID uint) (*dto.SSHKeyDownloadResponse, error) {
	key, err := s.repo.GetByID(keyID)
	if err != nil {
		return nil, fmt.Errorf("SSH Key 不存在")
	}

	privateKey, err := crypto.Decrypt(key.PrivateKey, s.masterKey)
	if err != nil {
		return nil, fmt.Errorf("解密私钥失败: %w", err)
	}

	return &dto.SSHKeyDownloadResponse{
		Name:       key.Name,
		PublicKey:  key.PublicKey,
		PrivateKey: privateKey,
	}, nil
}

// ListDeployLogs 查询部署日志
func (s *SSHKeyService) ListDeployLogs(keyID uint, page, pageSize int) ([]dto.SSHKeyDeployLogResponse, int64, error) {
	logs, total, err := s.repo.ListDeployLogsByKeyID(keyID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.SSHKeyDeployLogResponse, len(logs))
	for i, l := range logs {
		result[i] = dto.SSHKeyDeployLogResponse{
			ID:         l.ID,
			SSHKeyID:   l.SSHKeyID,
			SSHKeyName: l.SSHKeyName,
			AssetID:    l.AssetID,
			Hostname:   l.Hostname,
			IP:         l.IP,
			Username:   l.Username,
			Action:     l.Action,
			Status:     l.Status,
			Error:      l.Error,
			OperatorID: l.OperatorID,
			CreatedAt:  l.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return result, total, nil
}

// ListBindings 查询绑定关系
func (s *SSHKeyService) ListBindings(keyID uint) ([]dto.SSHKeyBindingResponse, error) {
	bindings, err := s.repo.ListBindingsByKeyID(keyID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.SSHKeyBindingResponse, len(bindings))
	for i, b := range bindings {
		result[i] = dto.SSHKeyBindingResponse{
			ID:       b.ID,
			SSHKeyID: b.SSHKeyID,
			AssetID:  b.AssetID,
			Hostname: b.Asset.Hostname,
			IP:       b.Asset.IP,
			Username: b.Username,
			Status:   b.Status,
		}
		if b.DeployedAt != nil {
			result[i].DeployedAt = b.DeployedAt.Format("2006-01-02 15:04:05")
		}
	}
	return result, nil
}
