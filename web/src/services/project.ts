import request from '../utils/request'
import type { PageResult } from './audit'

export interface ProjectItem {
  id: number
  name: string
  code: string
  description: string
  owner: string
  status: string
  asset_count: number
  created_at: string
  updated_at: string
}

export interface ProjectSimple {
  id: number
  name: string
  code: string
}

export interface ProjectListParams {
  keyword?: string
  status?: string
  page?: number
  page_size?: number
}

export interface CreateProjectParams {
  name: string
  code: string
  description?: string
  owner?: string
}

export interface UpdateProjectParams {
  name?: string
  description?: string
  owner?: string
  status?: string
}

export function getProjects(params?: ProjectListParams): Promise<PageResult<ProjectItem>> {
  return request.get('/projects', { params })
}

export function getProject(id: number): Promise<ProjectItem> {
  return request.get(`/projects/${id}`)
}

export function getAllProjects(): Promise<ProjectSimple[]> {
  return request.get('/projects/all')
}

export function createProject(data: CreateProjectParams): Promise<void> {
  return request.post('/projects', data)
}

export function updateProject(id: number, data: UpdateProjectParams): Promise<void> {
  return request.put(`/projects/${id}`, data)
}

export function deleteProject(id: number): Promise<void> {
  return request.delete(`/projects/${id}`)
}

// 项目资产指纹

export interface ProjectAssetStats {
  total_assets: number
  total_cpu: number
  total_memory: number
  total_disk: number
  probed: number
  running: number
  stopped: number
}

export interface DistributionItem {
  label: string
  value: number
}

export interface ProjectSummary {
  project: ProjectItem
  asset_stats: ProjectAssetStats
  status_distribution: DistributionItem[]
  type_distribution: DistributionItem[]
}

export function getProjectSummary(id: number): Promise<ProjectSummary> {
  return request.get(`/projects/${id}/summary`)
}

export function getProjectAssets(id: number, params?: { page?: number; page_size?: number }): Promise<PageResult<import('./asset').AssetItem>> {
  return request.get(`/projects/${id}/assets`, { params })
}
