import request from '../utils/request'
import type { PageResult } from './audit'

export interface TaskItem {
  id: number
  name: string
  type: string
  command: string
  status: string
  created_by: number
  created_by_name: string
  total_count: number
  success_count: number
  fail_count: number
  started_at: string
  finished_at: string
  created_at: string
}

export interface TaskResultItem {
  id: number
  task_id: number
  asset_id: number
  hostname: string
  ip: string
  status: string
  stdout: string
  stderr: string
  exit_code: number
  duration: number
}

export interface ExecuteTaskParams {
  name: string
  command: string
  asset_ids: number[]
}

export function executeTask(data: ExecuteTaskParams): Promise<TaskItem> {
  return request.post('/tasks/execute', data)
}

export function getTasks(params?: { status?: string; page?: number; page_size?: number }): Promise<PageResult<TaskItem>> {
  return request.get('/tasks', { params })
}

export function getTask(id: number): Promise<TaskItem> {
  return request.get(`/tasks/${id}`)
}

export function getTaskResults(id: number): Promise<TaskResultItem[]> {
  return request.get(`/tasks/${id}/results`)
}
