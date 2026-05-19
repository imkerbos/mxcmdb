import request from '../utils/request'

export interface PlaybookStep {
  name: string
  script: string
  timeout: number
  on_error: string // stop / continue
}

export interface Playbook {
  id: number
  name: string
  description: string
  category: string
  is_builtin: boolean
  enabled: boolean
  steps: PlaybookStep[]
  created_by: number
  created_at: string
  updated_at: string
}

export interface CreatePlaybookParams {
  name: string
  description: string
  steps: PlaybookStep[]
}

export interface UpdatePlaybookParams {
  name: string
  description: string
  enabled?: boolean
  steps: PlaybookStep[]
}

export interface ExecutePlaybookParams {
  playbook_ids: number[]
  asset_ids: number[]
}

export interface PlaybookJobResult {
  job_id: string
  execution_id: number
  total: number
}

export interface StepExecResult {
  name: string
  status: string
  output?: string
  error?: string
  latency: number
}

export interface PlaybookAssetResult {
  asset_id: number
  hostname: string
  ip: string
  status: string
  step_results: StepExecResult[]
}

export interface ExecutionListItem {
  id: number
  playbook_id: number
  playbook_name: string
  status: string
  total_assets: number
  success: number
  failed: number
  started_at: string
  finished_at: string
  created_by: number
}

export function getPlaybooks(): Promise<Playbook[]> {
  return request.get('/playbooks')
}

export function getPlaybook(id: number): Promise<Playbook> {
  return request.get(`/playbooks/${id}`)
}

export function createPlaybook(data: CreatePlaybookParams): Promise<Playbook> {
  return request.post('/playbooks', data)
}

export function updatePlaybook(id: number, data: UpdatePlaybookParams): Promise<void> {
  return request.put(`/playbooks/${id}`, data)
}

export function deletePlaybook(id: number): Promise<void> {
  return request.delete(`/playbooks/${id}`)
}

export function executePlaybook(data: ExecutePlaybookParams): Promise<PlaybookJobResult> {
  return request.post('/playbooks/execute', data)
}

export function getExecutions(params?: { page?: number; page_size?: number }): Promise<{
  list: ExecutionListItem[]
  total: number
  page: number
  page_size: number
}> {
  return request.get('/playbooks/executions', { params })
}

export function getExecutionResults(executionId: number): Promise<PlaybookAssetResult[]> {
  return request.get(`/playbooks/executions/${executionId}/results`)
}
