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
	"github.com/imkerbos/mxcmdb/internal/pkg/ws"
	"github.com/imkerbos/mxcmdb/internal/repository"
)

// taskAssetInfo 任务关联的资产信息
type taskAssetInfo struct {
	sshTask sshutil.Task
	result  *model.TaskResult
}

// TaskService 批量任务服务
type TaskService struct {
	taskRepo  *repository.TaskRepository
	assetRepo *repository.AssetRepository
	masterKey string
	configSvc *SystemConfigService
	hub       *ws.Hub
}

// NewTaskService 创建 TaskService
func NewTaskService(taskRepo *repository.TaskRepository, assetRepo *repository.AssetRepository, masterKey string, configSvc *SystemConfigService, hub *ws.Hub) *TaskService {
	return &TaskService{
		taskRepo:  taskRepo,
		assetRepo: assetRepo,
		masterKey: masterKey,
		configSvc: configSvc,
		hub:       hub,
	}
}

// Execute 创建并执行批量任务
func (s *TaskService) Execute(req dto.TaskExecuteRequest, userID uint) (*model.Task, error) {
	// 校验危险命令
	if err := s.checkDangerousCommand(req.Command); err != nil {
		return nil, err
	}

	// 创建任务
	now := time.Now()
	task := &model.Task{
		Name:       req.Name,
		Type:       "command",
		Command:    req.Command,
		Status:     "running",
		CreatedBy:  userID,
		TotalCount: len(req.AssetIDs),
		StartedAt:  &now,
	}
	if err := s.taskRepo.Create(task); err != nil {
		return nil, fmt.Errorf("创建任务失败: %w", err)
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

	// 构建 SSH 任务列表和结果记录
	var pendingResults []*model.TaskResult
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
		tr := &model.TaskResult{
			TaskID:   task.ID,
			AssetID:  asset.ID,
			Hostname: asset.Hostname,
			IP:       asset.IP,
			Status:   "pending",
		}
		pendingResults = append(pendingResults, tr)
		pendingTasks = append(pendingTasks, st)
	}

	// 批量插入结果记录
	if err := s.taskRepo.CreateResults(pendingResults); err != nil {
		return nil, fmt.Errorf("创建任务结果失败: %w", err)
	}

	var infos []taskAssetInfo
	var sshTasks []sshutil.Task
	for i, tr := range pendingResults {
		sshTasks = append(sshTasks, pendingTasks[i])
		infos = append(infos, taskAssetInfo{sshTask: pendingTasks[i], result: tr})
	}

	// 异步执行
	go s.runTask(task, infos, sshTasks)

	return task, nil
}

// runTask 异步执行批量任务
func (s *TaskService) runTask(task *model.Task, infos []taskAssetInfo, sshTasks []sshutil.Task) {
	timeout := time.Duration(s.configSvc.GetInt("ssh.timeout", 30)) * time.Second
	maxConcurrent := s.configSvc.GetInt("ssh.max_concurrent", 100)
	pool := sshutil.NewWorkerPool(maxConcurrent)

	roomID := fmt.Sprintf("task_%d", task.ID)
	successCount := 0
	failCount := 0

	results := pool.Execute(sshTasks, func(t sshutil.Task) sshutil.ExecResult {
		client, err := sshutil.Dial(sshutil.ClientConfig{
			Host: t.Host, Port: t.Port, User: t.User,
			Password: t.Password, Timeout: timeout,
		})
		if err != nil {
			return sshutil.ExecResult{Err: err}
		}
		defer client.Close()
		return sshutil.RunCommand(client, task.Command, timeout*2)
	})

	for i, r := range results {
		if i >= len(infos) {
			break
		}
		tr := infos[i].result
		tr.Stdout = r.Result.Stdout
		tr.Stderr = r.Result.Stderr
		tr.ExitCode = r.Result.ExitCode
		tr.Duration = r.Result.Duration.Milliseconds()

		if r.Result.Err != nil {
			tr.Status = "failed"
			tr.Stderr = r.Result.Err.Error()
			failCount++
		} else if r.Result.ExitCode != 0 {
			tr.Status = "failed"
			failCount++
		} else {
			tr.Status = "success"
			successCount++
		}

		_ = s.taskRepo.UpdateResult(tr)

		// WebSocket 推送单条结果
		s.hub.Broadcast(roomID, ws.Message{
			Type: "task_result",
			Data: dto.TaskResultResponse{
				ID:       tr.ID,
				TaskID:   tr.TaskID,
				AssetID:  tr.AssetID,
				Hostname: tr.Hostname,
				IP:       tr.IP,
				Status:   tr.Status,
				Stdout:   tr.Stdout,
				Stderr:   tr.Stderr,
				ExitCode: tr.ExitCode,
				Duration: tr.Duration,
			},
		})
	}

	// 更新任务状态
	now := time.Now()
	task.SuccessCount = successCount
	task.FailCount = failCount
	task.FinishedAt = &now
	if failCount == 0 {
		task.Status = "completed"
	} else if successCount == 0 {
		task.Status = "failed"
	} else {
		task.Status = "completed"
	}
	_ = s.taskRepo.Update(task)

	// 推送任务完成
	s.hub.Broadcast(roomID, ws.Message{
		Type: "task_done",
		Data: map[string]interface{}{
			"task_id":       task.ID,
			"status":        task.Status,
			"success_count": successCount,
			"fail_count":    failCount,
		},
	})

	logger.Log.Infof("批量任务 %d 完成: success=%d, fail=%d", task.ID, successCount, failCount)
}

// checkDangerousCommand 检查危险命令
func (s *TaskService) checkDangerousCommand(command string) error {
	dangerousStr := s.configSvc.GetWithDefault("security.dangerous_commands", "rm -rf /,shutdown,reboot,mkfs,dd if=,halt,poweroff")
	dangerous := strings.Split(dangerousStr, ",")
	cmdLower := strings.ToLower(command)
	for _, d := range dangerous {
		d = strings.TrimSpace(strings.ToLower(d))
		if d != "" && strings.Contains(cmdLower, d) {
			return fmt.Errorf("命令包含危险操作: %s", d)
		}
	}
	return nil
}

// List 任务列表
func (s *TaskService) List(status string, page, pageSize int) ([]dto.TaskResponse, int64, error) {
	tasks, total, err := s.taskRepo.List(status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.TaskResponse, len(tasks))
	for i, t := range tasks {
		result[i] = toTaskResponse(&t)
	}
	return result, total, nil
}

// GetByID 获取任务详情
func (s *TaskService) GetByID(id uint) (*dto.TaskResponse, error) {
	task, err := s.taskRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	resp := toTaskResponse(task)
	return &resp, nil
}

// GetResults 获取任务结果
func (s *TaskService) GetResults(taskID uint) ([]dto.TaskResultResponse, error) {
	results, err := s.taskRepo.ListResults(taskID)
	if err != nil {
		return nil, err
	}
	resp := make([]dto.TaskResultResponse, len(results))
	for i, r := range results {
		resp[i] = dto.TaskResultResponse{
			ID:       r.ID,
			TaskID:   r.TaskID,
			AssetID:  r.AssetID,
			Hostname: r.Hostname,
			IP:       r.IP,
			Status:   r.Status,
			Stdout:   r.Stdout,
			Stderr:   r.Stderr,
			ExitCode: r.ExitCode,
			Duration: r.Duration,
		}
	}
	return resp, nil
}

func toTaskResponse(t *model.Task) dto.TaskResponse {
	resp := dto.TaskResponse{
		ID:           t.ID,
		Name:         t.Name,
		Type:         t.Type,
		Command:      t.Command,
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
