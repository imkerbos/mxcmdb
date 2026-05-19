package service

import (
	"fmt"
	"os"
	"time"

	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/pkg/crypto"
	"github.com/imkerbos/mxcmdb/internal/pkg/logger"
	"github.com/imkerbos/mxcmdb/internal/pkg/sshutil"
	"github.com/imkerbos/mxcmdb/internal/repository"
)

// FileDistributionService 文件分发服务
type FileDistributionService struct {
	fileTaskRepo *repository.FileTaskRepository
	assetRepo    *repository.AssetRepository
	masterKey    string
	configSvc    *SystemConfigService
}

// NewFileDistributionService 创建 FileDistributionService
func NewFileDistributionService(fileTaskRepo *repository.FileTaskRepository, assetRepo *repository.AssetRepository, masterKey string, configSvc *SystemConfigService) *FileDistributionService {
	return &FileDistributionService{
		fileTaskRepo: fileTaskRepo,
		assetRepo:    assetRepo,
		masterKey:    masterKey,
		configSvc:    configSvc,
	}
}

// Distribute 文件分发
func (s *FileDistributionService) Distribute(localPath, fileName string, fileSize int64, req dto.FileDistributeRequest, userID uint) (*model.FileTask, error) {
	fileMode := req.FileMode
	if fileMode == "" {
		fileMode = "0644"
	}

	now := time.Now()
	task := &model.FileTask{
		Name:       req.Name,
		FileName:   fileName,
		FileSize:   fileSize,
		RemotePath: req.RemotePath,
		FileMode:   fileMode,
		Status:     "running",
		CreatedBy:  userID,
		TotalCount: len(req.AssetIDs),
		StartedAt:  &now,
	}
	if err := s.fileTaskRepo.Create(task); err != nil {
		return nil, fmt.Errorf("创建文件任务失败: %w", err)
	}

	type assetTarget struct {
		sshTask sshutil.Task
		result  *model.FileTaskResult
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

	var pendingResults []*model.FileTaskResult
	var pendingTasks []sshutil.Task
	for _, assetID := range req.AssetIDs {
		asset, ok := assetMap[assetID]
		if !ok {
			continue
		}
		password := ""
		if asset.SshPassword != "" {
			password, _ = crypto.Decrypt(asset.SshPassword, s.masterKey)
		}
		st := sshutil.Task{
			Host: asset.IP, Port: asset.Port, User: asset.SshUser,
			Password: password,
		}
		tr := &model.FileTaskResult{
			FileTaskID: task.ID,
			AssetID:    asset.ID,
			Hostname:   asset.Hostname,
			IP:         asset.IP,
			Status:     "pending",
		}
		pendingResults = append(pendingResults, tr)
		pendingTasks = append(pendingTasks, st)
	}

	// 批量插入结果记录
	_ = s.fileTaskRepo.CreateResults(pendingResults)

	var targets []assetTarget
	var sshTasks []sshutil.Task
	for i, tr := range pendingResults {
		sshTasks = append(sshTasks, pendingTasks[i])
		targets = append(targets, assetTarget{sshTask: pendingTasks[i], result: tr})
	}

	// 异步分发
	go func() {
		defer os.Remove(localPath)

		timeout := time.Duration(s.configSvc.GetInt("ssh.timeout", 30)) * time.Second
		maxConcurrent := s.configSvc.GetInt("ssh.max_concurrent", 100)
		pool := sshutil.NewWorkerPool(maxConcurrent)

		// 解析文件权限
		var mode os.FileMode = 0644
		if fileMode != "" {
			var m uint32
			if _, err := fmt.Sscanf(fileMode, "%o", &m); err == nil {
				mode = os.FileMode(m)
			}
		}

		successCount := 0
		failCount := 0

		results := pool.Execute(sshTasks, func(t sshutil.Task) sshutil.ExecResult {
			start := time.Now()
			client, err := sshutil.Dial(sshutil.ClientConfig{
				Host: t.Host, Port: t.Port, User: t.User,
				Password: t.Password, Timeout: timeout,
			})
			if err != nil {
				return sshutil.ExecResult{Err: err, Duration: time.Since(start)}
			}
			defer client.Close()

			if err := sshutil.SFTPUpload(client, localPath, req.RemotePath, mode); err != nil {
				return sshutil.ExecResult{Err: err, Duration: time.Since(start)}
			}
			return sshutil.ExecResult{Duration: time.Since(start)}
		})

		for i, r := range results {
			if i >= len(targets) {
				break
			}
			tr := targets[i].result
			tr.Duration = r.Result.Duration.Milliseconds()
			if r.Result.Err != nil {
				tr.Status = "failed"
				tr.ErrorMsg = r.Result.Err.Error()
				failCount++
			} else {
				tr.Status = "success"
				successCount++
			}
			_ = s.fileTaskRepo.UpdateResult(tr)
		}

		finishNow := time.Now()
		task.SuccessCount = successCount
		task.FailCount = failCount
		task.FinishedAt = &finishNow
		if failCount == 0 {
			task.Status = "completed"
		} else if successCount == 0 {
			task.Status = "failed"
		} else {
			task.Status = "completed"
		}
		_ = s.fileTaskRepo.Update(task)
		logger.Log.Infof("文件分发任务 %d 完成: success=%d, fail=%d", task.ID, successCount, failCount)
	}()

	return task, nil
}

// List 文件任务列表
func (s *FileDistributionService) List(page, pageSize int) ([]dto.FileTaskResponse, int64, error) {
	tasks, total, err := s.fileTaskRepo.List(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.FileTaskResponse, len(tasks))
	for i, t := range tasks {
		result[i] = toFileTaskResponse(&t)
	}
	return result, total, nil
}

// GetResults 查询结果
func (s *FileDistributionService) GetResults(taskID uint) ([]dto.FileTaskResultResponse, error) {
	results, err := s.fileTaskRepo.ListResults(taskID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.FileTaskResultResponse, len(results))
	for i, r := range results {
		resp[i] = dto.FileTaskResultResponse{
			ID:         r.ID,
			FileTaskID: r.FileTaskID,
			AssetID:    r.AssetID,
			Hostname:   r.Hostname,
			IP:         r.IP,
			Status:     r.Status,
			ErrorMsg:   r.ErrorMsg,
			Duration:   r.Duration,
		}
	}
	return resp, nil
}

func toFileTaskResponse(t *model.FileTask) dto.FileTaskResponse {
	resp := dto.FileTaskResponse{
		ID:           t.ID,
		Name:         t.Name,
		FileName:     t.FileName,
		FileSize:     t.FileSize,
		RemotePath:   t.RemotePath,
		FileMode:     t.FileMode,
		Status:       t.Status,
		CreatedBy:    t.CreatedBy,
		TotalCount:   t.TotalCount,
		SuccessCount: t.SuccessCount,
		FailCount:    t.FailCount,
		CreatedAt:    t.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if t.StartedAt != nil {
		resp.StartedAt = t.StartedAt.Format("2006-01-02 15:04:05")
	}
	if t.FinishedAt != nil {
		resp.FinishedAt = t.FinishedAt.Format("2006-01-02 15:04:05")
	}
	return resp
}
