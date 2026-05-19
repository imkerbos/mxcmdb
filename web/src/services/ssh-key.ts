import request from '../utils/request'
import type { PageResult } from './audit'

export interface SSHKeyItem {
  id: number
  name: string
  public_key: string
  fingerprint: string
  key_type: string
  comment: string
  created_at: string
}

export interface SSHKeyBinding {
  id: number
  ssh_key_id: number
  asset_id: number
  hostname: string
  ip: string
  username: string
  status: string
  deployed_at: string
}

export interface CreateSSHKeyParams {
  name: string
  public_key?: string
  private_key?: string
  key_type?: string
  comment?: string
}

export interface DeploySSHKeyParams {
  asset_ids: number[]
  username: string
}

export function getSSHKeys(params?: { keyword?: string; page?: number; page_size?: number }): Promise<PageResult<SSHKeyItem>> {
  return request.get('/sshkeys', { params })
}

export async function getAllSSHKeys(): Promise<SSHKeyItem[]> {
  const res: PageResult<SSHKeyItem> = await request.get('/sshkeys', { params: { page: 1, page_size: 200 } })
  return res.list || []
}

export function createSSHKey(data: CreateSSHKeyParams): Promise<SSHKeyItem> {
  return request.post('/sshkeys', data)
}

export function deleteSSHKey(id: number): Promise<void> {
  return request.delete(`/sshkeys/${id}`)
}

export function deploySSHKey(id: number, data: DeploySSHKeyParams): Promise<void> {
  return request.post(`/sshkeys/${id}/deploy`, data)
}

export function getSSHKeyBindings(id: number): Promise<SSHKeyBinding[]> {
  return request.get(`/sshkeys/${id}/bindings`)
}

// 撤销（回收）SSH Key
export interface RevokeSSHKeyParams {
  asset_ids: number[]
  username: string
}

export interface RevokeResultItem {
  asset_id: number
  hostname: string
  ip: string
  status: string
  error?: string
}

export interface RevokeResponse {
  total: number
  success: number
  failed: number
  results: RevokeResultItem[]
}

export function revokeSSHKey(id: number, data: RevokeSSHKeyParams): Promise<RevokeResponse> {
  return request.post(`/sshkeys/${id}/revoke`, data)
}

// SSH Key 轮换
export interface RotateSSHKeyParams {
  new_key_id?: number
  new_name?: string
  new_comment?: string
}

export interface RotateResponse {
  old_key_id: number
  new_key: SSHKeyItem
  deploy: RevokeResponse | null
  revoke: RevokeResponse | null
}

export function rotateSSHKey(id: number, data: RotateSSHKeyParams): Promise<RotateResponse> {
  return request.post(`/sshkeys/${id}/rotate`, data)
}

// 离职清理
export interface DepartureCleanupParams {
  key_ids: number[]
  asset_ids?: number[]
}

export interface DepartureCleanupResponse {
  total_keys: number
  total_assets: number
  revoke_results: RevokeResponse[]
}

export function departureCleanup(data: DepartureCleanupParams): Promise<DepartureCleanupResponse> {
  return request.post('/sshkeys/departure-cleanup', data)
}

// 下载 SSH Key（含解密私钥）
export interface SSHKeyDownloadResponse {
  name: string
  public_key: string
  private_key: string
}

export function downloadSSHKey(id: number): Promise<SSHKeyDownloadResponse> {
  return request.get(`/sshkeys/${id}/download`)
}

// 部署日志
export interface SSHKeyDeployLog {
  id: number
  ssh_key_id: number
  ssh_key_name: string
  asset_id: number
  hostname: string
  ip: string
  username: string
  action: string
  status: string
  error?: string
  operator_id: number
  created_at: string
}

export function getSSHKeyDeployLogs(id: number, params?: { page?: number; page_size?: number }): Promise<PageResult<SSHKeyDeployLog>> {
  return request.get(`/sshkeys/${id}/deploy-logs`, { params })
}
