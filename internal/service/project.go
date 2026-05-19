package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/imkerbos/mxcmdb/internal/dto"
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/repository"
)

// ProjectService 项目管理服务
type ProjectService struct {
	repo      *repository.ProjectRepository
	probeRepo *repository.ProbeResultRepository
}

// NewProjectService 创建 ProjectService
func NewProjectService(repo *repository.ProjectRepository, probeRepo *repository.ProbeResultRepository) *ProjectService {
	return &ProjectService{repo: repo, probeRepo: probeRepo}
}

// List 项目列表
func (s *ProjectService) List(keyword, status string, page, pageSize int) ([]dto.ProjectResponse, int64, error) {
	projects, total, err := s.repo.List(keyword, status, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("查询项目列表失败: %w", err)
	}

	result := make([]dto.ProjectResponse, len(projects))
	for i, p := range projects {
		assetCount, _ := s.repo.CountAssets(p.ID)
		result[i] = dto.ProjectResponse{
			ID:          p.ID,
			Name:        p.Name,
			Code:        p.Code,
			Description: p.Description,
			Owner:       p.Owner,
			Status:      p.Status,
			AssetCount:  assetCount,
			CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return result, total, nil
}

// GetByID 项目详情
func (s *ProjectService) GetByID(id uint) (*dto.ProjectResponse, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("项目不存在")
	}
	assetCount, _ := s.repo.CountAssets(p.ID)
	resp := &dto.ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Code:        p.Code,
		Description: p.Description,
		Owner:       p.Owner,
		Status:      p.Status,
		AssetCount:  assetCount,
		CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	return resp, nil
}

// Create 创建项目
func (s *ProjectService) Create(req dto.CreateProjectRequest) error {
	if existing, _ := s.repo.GetByCode(req.Code); existing != nil {
		return fmt.Errorf("项目代码 %s 已存在", req.Code)
	}
	if existing, _ := s.repo.GetByName(req.Name); existing != nil {
		return fmt.Errorf("项目名称 %s 已存在", req.Name)
	}

	project := &model.Project{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Owner:       req.Owner,
		Status:      "active",
	}
	return s.repo.Create(project)
}

// Update 更新项目
func (s *ProjectService) Update(id uint, req dto.UpdateProjectRequest) error {
	project, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("项目不存在")
	}

	if req.Name != "" && req.Name != project.Name {
		if existing, _ := s.repo.GetByName(req.Name); existing != nil && existing.ID != id {
			return fmt.Errorf("项目名称 %s 已存在", req.Name)
		}
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.Owner != "" {
		project.Owner = req.Owner
	}
	if req.Status != "" {
		project.Status = req.Status
	}

	return s.repo.Update(project)
}

// Delete 删除项目
func (s *ProjectService) Delete(id uint) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return fmt.Errorf("项目不存在")
	}
	assetCount, _ := s.repo.CountAssets(id)
	if assetCount > 0 {
		return fmt.Errorf("项目下仍有 %d 个资产，请先解除关联", assetCount)
	}
	return s.repo.Delete(id)
}

// ListAll 获取所有项目（用于选择器）
func (s *ProjectService) ListAll() ([]dto.ProjectSimple, error) {
	projects, err := s.repo.ListAll()
	if err != nil {
		return nil, fmt.Errorf("查询项目列表失败: %w", err)
	}
	result := make([]dto.ProjectSimple, len(projects))
	for i, p := range projects {
		result[i] = dto.ProjectSimple{ID: p.ID, Name: p.Name, Code: p.Code}
	}
	return result, nil
}

// GetSummary 获取项目资产指纹聚合
func (s *ProjectService) GetSummary(id uint) (*dto.ProjectSummary, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("项目不存在")
	}

	// 项目基本信息
	assetCount, _ := s.repo.CountAssets(p.ID)
	projectResp := dto.ProjectResponse{
		ID: p.ID, Name: p.Name, Code: p.Code,
		Description: p.Description, Owner: p.Owner, Status: p.Status,
		AssetCount: assetCount,
		CreatedAt:  p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// 获取项目下资产 ID
	assetIDs, err := s.repo.ListAssetIDsByProjectID(id)
	if err != nil || len(assetIDs) == 0 {
		return &dto.ProjectSummary{Project: projectResp}, nil
	}

	// 获取资产列表统计
	assets, _, err := s.repo.ListAssetsByProjectID(id, 1, 10000)
	if err != nil {
		assets = nil
	}

	stats := dto.ProjectAssetStats{TotalAssets: int64(len(assets))}
	statusDist := map[string]int64{}
	typeDist := map[string]int64{}

	for _, a := range assets {
		st := a.Status
		if st == "" {
			st = "unknown"
		}
		statusDist[st]++
		if st == "running" {
			stats.Running++
		} else if st == "stopped" {
			stats.Stopped++
		}
		if a.ProbeLastAt != nil {
			stats.Probed++
		}
		tp := a.Type
		if tp == "" {
			tp = "unknown"
		}
		typeDist[tp]++
	}

	// 获取探针数据聚合 CPU/Memory/Disk
	probeResults, _ := s.probeRepo.GetLatestByAssetIDs(assetIDs)
	for _, pr := range probeResults {
		stats.TotalCPU += parseCPUCores(pr.CPU)
		stats.TotalMemory += parseMemoryMB(pr.Memory)
		stats.TotalDisk += parseDiskGB(pr.Disk)
	}

	// 构建分布数组
	var statusItems []dto.DistributionItem
	for k, v := range statusDist {
		statusItems = append(statusItems, dto.DistributionItem{Label: k, Value: v})
	}
	var typeItems []dto.DistributionItem
	for k, v := range typeDist {
		typeItems = append(typeItems, dto.DistributionItem{Label: k, Value: v})
	}

	return &dto.ProjectSummary{
		Project:    projectResp,
		AssetStats: stats,
		StatusDist: statusItems,
		TypeDist:   typeItems,
	}, nil
}

// ListAssets 获取项目下资产列表
func (s *ProjectService) ListAssets(id uint, page, pageSize int) ([]model.Asset, int64, error) {
	if _, err := s.repo.GetByID(id); err != nil {
		return nil, 0, fmt.Errorf("项目不存在")
	}
	return s.repo.ListAssetsByProjectID(id, page, pageSize)
}

// parseCPUCores 从 lscpu 输出解析 CPU 核心数
func parseCPUCores(cpuText string) int64 {
	// lscpu 格式: "CPU(s):                4"
	re := regexp.MustCompile(`(?i)CPU\(s\):\s*(\d+)`)
	m := re.FindStringSubmatch(cpuText)
	if len(m) >= 2 {
		v, _ := strconv.ParseInt(m[1], 10, 64)
		return v
	}
	// /proc/cpuinfo 回退: 计算 "processor" 行数
	count := int64(0)
	for _, line := range strings.Split(cpuText, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "processor") {
			count++
		}
	}
	return count
}

// parseMemoryMB 从 free -m 输出解析总内存 (MB)
func parseMemoryMB(memText string) int64 {
	// free -m 格式:
	//               total        used        free ...
	// Mem:           7982        1234        4567 ...
	for _, line := range strings.Split(memText, "\n") {
		if strings.HasPrefix(line, "Mem:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				v, _ := strconv.ParseInt(fields[1], 10, 64)
				return v
			}
		}
	}
	return 0
}

// parseDiskGB 从 df -h 输出解析总磁盘 (GB)
func parseDiskGB(diskText string) int64 {
	// df -h 格式:
	// Filesystem      Size  Used Avail Use% Mounted on
	// /dev/sda1       100G   50G   50G  50% /
	var totalGB float64
	seen := map[string]bool{}
	for _, line := range strings.Split(diskText, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		fs := fields[0]
		// 跳过表头和非磁盘文件系统
		if fs == "Filesystem" || strings.HasPrefix(fs, "tmpfs") || strings.HasPrefix(fs, "devtmpfs") || fs == "overlay" || fs == "shm" {
			continue
		}
		if seen[fs] {
			continue
		}
		seen[fs] = true
		totalGB += parseSizeToGB(fields[1])
	}
	return int64(totalGB)
}

// parseSizeToGB 解析 "100G" / "500M" / "2T" 为 GB
func parseSizeToGB(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	v, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}
	switch unit {
	case 'T', 't':
		return v * 1024
	case 'G', 'g':
		return v
	case 'M', 'm':
		return v / 1024
	case 'K', 'k':
		return v / (1024 * 1024)
	default:
		return 0
	}
}
