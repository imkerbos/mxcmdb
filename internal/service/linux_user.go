package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/pkg/crypto"
	"github.com/imkerbos/mxcmdb/internal/pkg/logger"
	"github.com/imkerbos/mxcmdb/internal/pkg/sshutil"
	"github.com/imkerbos/mxcmdb/internal/repository"
)

// LinuxUserService Linux 用户管理服务
type LinuxUserService struct {
	repo      *repository.LinuxUserRepository
	assetRepo *repository.AssetRepository
	masterKey string
	configSvc *SystemConfigService
}

// NewLinuxUserService 创建 LinuxUserService
func NewLinuxUserService(repo *repository.LinuxUserRepository, assetRepo *repository.AssetRepository, masterKey string, configSvc *SystemConfigService) *LinuxUserService {
	return &LinuxUserService{
		repo:      repo,
		assetRepo: assetRepo,
		masterKey: masterKey,
		configSvc: configSvc,
	}
}

// Create 批量创建 Linux 用户
func (s *LinuxUserService) Create(req dto.CreateLinuxUserRequest) error {
	timeout := time.Duration(s.configSvc.GetInt("ssh.timeout", 30)) * time.Second
	maxConcurrent := s.configSvc.GetInt("ssh.max_concurrent", 100)
	pool := sshutil.NewWorkerPool(maxConcurrent)

	shell := req.Shell
	if shell == "" {
		shell = "/bin/bash"
	}

	// 构建命令（非 root 用户需要 sudo）
	cmd := fmt.Sprintf("sudo useradd -m -s %s %s", shell, req.Username)
	if req.Sudo {
		cmd += fmt.Sprintf(" && sudo usermod -aG sudo %s 2>/dev/null || sudo usermod -aG wheel %s 2>/dev/null", req.Username, req.Username)
	}
	cmd += fmt.Sprintf(" && id %s", req.Username)

	// 批量加载资产
	assetList, err := s.assetRepo.GetByIDs(req.AssetIDs)
	if err != nil {
		return fmt.Errorf("加载资产失败: %w", err)
	}
	assetMap := make(map[uint]*model.Asset, len(assetList))
	for i := range assetList {
		assetMap[assetList[i].ID] = &assetList[i]
	}

	var tasks []sshutil.Task
	var assets []*model.Asset
	for _, assetID := range req.AssetIDs {
		asset, ok := assetMap[assetID]
		if !ok {
			continue
		}
		password := ""
		if asset.SshPassword != "" {
			password, _ = crypto.Decrypt(asset.SshPassword, s.masterKey)
		}
		tasks = append(tasks, sshutil.Task{
			Host: asset.IP, Port: asset.Port, User: asset.SshUser,
			Password: password,
		})
		assets = append(assets, asset)
	}

	go func() {
		results := pool.Execute(tasks, func(t sshutil.Task) sshutil.ExecResult {
			client, err := sshutil.Dial(sshutil.ClientConfig{
				Host: t.Host, Port: t.Port, User: t.User,
				Password: t.Password, Timeout: timeout,
			})
			if err != nil {
				return sshutil.ExecResult{Err: err}
			}
			defer client.Close()
			return sshutil.RunCommand(client, cmd, timeout)
		})

		for i, r := range results {
			if i >= len(assets) {
				break
			}
			status := "active"
			if r.Result.Err != nil || r.Result.ExitCode != 0 {
				status = "failed"
			}

			// 解析 uid/gid
			uid, gid := parseIDOutput(r.Result.Stdout)

			user := &model.LinuxUser{
				Username: req.Username,
				AssetID:  assets[i].ID,
				UID:      uid,
				GID:      gid,
				Home:     fmt.Sprintf("/home/%s", req.Username),
				Shell:    shell,
				Sudo:     req.Sudo,
				Status:   status,
			}
			if err := s.repo.Create(user); err != nil {
				logger.Log.Errorf("保存 Linux 用户记录失败: %v", err)
			}
		}
		logger.Log.Infof("Linux 用户 %s 创建完成，%d 台主机", req.Username, len(results))
	}()

	return nil
}

// Delete 批量删除 Linux 用户
func (s *LinuxUserService) Delete(username string, req dto.DeleteLinuxUserRequest) error {
	timeout := time.Duration(s.configSvc.GetInt("ssh.timeout", 30)) * time.Second
	maxConcurrent := s.configSvc.GetInt("ssh.max_concurrent", 100)
	pool := sshutil.NewWorkerPool(maxConcurrent)

	cmd := fmt.Sprintf("sudo userdel -r %s 2>/dev/null; echo $?", username)

	// 批量加载资产
	assetList, err := s.assetRepo.GetByIDs(req.AssetIDs)
	if err != nil {
		return fmt.Errorf("加载资产失败: %w", err)
	}
	delAssetMap := make(map[uint]*model.Asset, len(assetList))
	for i := range assetList {
		delAssetMap[assetList[i].ID] = &assetList[i]
	}

	var tasks []sshutil.Task
	var assetIDs []uint
	for _, assetID := range req.AssetIDs {
		asset, ok := delAssetMap[assetID]
		if !ok {
			continue
		}
		password := ""
		if asset.SshPassword != "" {
			password, _ = crypto.Decrypt(asset.SshPassword, s.masterKey)
		}
		tasks = append(tasks, sshutil.Task{
			Host: asset.IP, Port: asset.Port, User: asset.SshUser,
			Password: password,
		})
		assetIDs = append(assetIDs, assetID)
	}

	go func() {
		pool.Execute(tasks, func(t sshutil.Task) sshutil.ExecResult {
			client, err := sshutil.Dial(sshutil.ClientConfig{
				Host: t.Host, Port: t.Port, User: t.User,
				Password: t.Password, Timeout: timeout,
			})
			if err != nil {
				return sshutil.ExecResult{Err: err}
			}
			defer client.Close()
			return sshutil.RunCommand(client, cmd, timeout)
		})

		// 更新本地记录
		for _, assetID := range assetIDs {
			existing, err := s.repo.FindByUsernameAndAsset(username, assetID)
			if err == nil {
				existing.Status = "deleted"
				_ = s.repo.Update(existing)
			}
		}
		logger.Log.Infof("Linux 用户 %s 删除完成，%d 台主机", username, len(tasks))
	}()

	return nil
}

// List Linux 用户列表
func (s *LinuxUserService) List(username string, page, pageSize int) ([]dto.LinuxUserResponse, int64, error) {
	users, total, err := s.repo.List(username, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.LinuxUserResponse, len(users))
	for i, u := range users {
		result[i] = dto.LinuxUserResponse{
			ID:       u.ID,
			Username: u.Username,
			AssetID:  u.AssetID,
			Hostname: u.Asset.Hostname,
			IP:       u.Asset.IP,
			UID:      u.UID,
			GID:      u.GID,
			Home:     u.Home,
			Shell:    u.Shell,
			Sudo:     u.Sudo,
			Status:   u.Status,
		}
	}
	return result, total, nil
}

// parseIDOutput 解析 id 命令输出获取 uid 和 gid
func parseIDOutput(output string) (int, int) {
	var uid, gid int
	// id output format: uid=1000(user) gid=1000(user) groups=...
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "uid=") {
			_, _ = fmt.Sscanf(line, "uid=%d", &uid)
			if idx := strings.Index(line, "gid="); idx > 0 {
				_, _ = fmt.Sscanf(line[idx:], "gid=%d", &gid)
			}
			break
		}
	}
	return uid, gid
}
