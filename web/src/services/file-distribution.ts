import request from '../utils/request'
import type { PageResult } from './audit'

export interface FileTaskItem {
  id: number
  name: string
  file_name: string
  file_size: number
  remote_path: string
  file_mode: string
  status: string
  created_by: number
  total_count: number
  success_count: number
  fail_count: number
  started_at: string
  finished_at: string
  created_at: string
}

export interface FileTaskResultItem {
  id: number
  file_task_id: number
  asset_id: number
  hostname: string
  ip: string
  status: string
  error_msg: string
  duration: number
}

export function distributeFile(formData: FormData): Promise<FileTaskItem> {
  return request.post('/files/distribute', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 60000,
  })
}

export function getFileTasks(params?: { page?: number; page_size?: number }): Promise<PageResult<FileTaskItem>> {
  return request.get('/files/tasks', { params })
}

export function getFileTaskResults(id: number): Promise<FileTaskResultItem[]> {
  return request.get(`/files/tasks/${id}/results`)
}
