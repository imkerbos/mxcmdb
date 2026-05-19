import request from '../utils/request'
import type { PageResult } from './audit'

export interface AssetTag {
  key: string
  value: string
}

export interface AssetItem {
  id: number
  hostname: string
  ip: string
  port: number
  os: string
  type: string
  source: string
  cloud_account_id: number | null
  instance_id: string
  region: string
  zone: string
  status: string
  spec: string
  department: string
  project_id: number | null
  project_name: string
  owner: string
  environment: string
  business_group: string
  ssh_user: string
  ssh_key_id: number | null
  ssh_key_name: string
  has_password: boolean
  probe_last_at: string
  tags: AssetTag[]
  created_at: string
  updated_at: string
}

export interface AssetListParams {
  keyword?: string
  type?: string
  source?: string
  status?: string
  environment?: string
  department?: string
  project_id?: number
  owner?: string
  page?: number
  page_size?: number
}

export interface CreateAssetParams {
  hostname?: string
  ip: string
  port?: number
  os?: string
  type: string
  source?: string
  status?: string
  spec?: string
  department?: string
  project_id?: number | null
  owner?: string
  environment?: string
  business_group?: string
  ssh_user?: string
  ssh_password?: string
  ssh_key_id?: number | null
  tags?: AssetTag[]
}

export interface UpdateAssetParams {
  hostname?: string
  port?: number
  os?: string
  type?: string
  status?: string
  spec?: string
  department?: string
  project_id?: number | null
  owner?: string
  environment?: string
  business_group?: string
  ssh_user?: string
  ssh_password?: string
  ssh_key_id?: number | null
  clear_ssh_key?: boolean
  tags?: AssetTag[]
}

export interface AssetStats {
  total_cloud: number
  total_idc: number
  total_online: number
  total_offline: number
  total_probed: number
}

export interface AssetSimple {
  id: number
  hostname: string
  ip: string
  type: string
  status: string
  environment: string
  project_id: number | null
}

export function getAssets(params?: AssetListParams): Promise<PageResult<AssetItem>> {
  return request.get('/assets', { params })
}

export function getAsset(id: number): Promise<AssetItem> {
  return request.get(`/assets/${id}`)
}

// 资产详情（聚合数据）

export interface AssetDetailProbe {
  hostname: string
  cpu: string
  memory: string
  disk: string
  network: string
  os: string
  kernel: string
  docker_version: string
  running_services: string
  ssh_users: string
  uptime: string
  dns: string
  gateway: string
  manufacturer: string
  product_model: string
  serial_number: string
  public_ip: string
  processes: string
  listeners: string
  collected_at: string
  status: string
}

export interface AssetDetailSSHKeyBinding {
  id: number
  ssh_key_id: number
  key_name: string
  username: string
  status: string
  deployed_at: string
}

export interface AssetDetailLinuxUser {
  id: number
  username: string
  uid: number
  gid: number
  home: string
  shell: string
  sudo: boolean
  status: string
}

export interface AssetDetailTerminalSession {
  id: number
  user_id: number
  username: string
  status: string
  client_ip: string
  started_at: string
  finished_at: string
}

export interface AssetDetailProbeHistory {
  id: number
  status: string
  os: string
  kernel: string
  collected_at: string
}

export interface AssetDetail extends AssetItem {
  probe: AssetDetailProbe | null
  ssh_key_bindings: AssetDetailSSHKeyBinding[]
  linux_users: AssetDetailLinuxUser[]
  terminal_sessions: AssetDetailTerminalSession[]
  probe_history: AssetDetailProbeHistory[]
}

export function getAssetDetail(id: number): Promise<AssetDetail> {
  return request.get(`/assets/${id}/detail`)
}

export function getAllAssets(): Promise<AssetSimple[]> {
  return request.get('/assets/all')
}

export function getAssetStats(): Promise<AssetStats> {
  return request.get('/assets/stats')
}

export function createAsset(data: CreateAssetParams): Promise<void> {
  return request.post('/assets', data)
}

export function updateAsset(id: number, data: UpdateAssetParams): Promise<void> {
  return request.put(`/assets/${id}`, data)
}

export function deleteAsset(id: number): Promise<void> {
  return request.delete(`/assets/${id}`)
}

// 批量操作

export interface BatchImportResult {
  total: number
  success: number
  failed: number
  errors?: { index: number; ip: string; message: string }[]
}

export interface TestConnectionResult {
  asset_id: number
  hostname: string
  ip: string
  status: string
  latency: number
  auth_method: string
  error?: string
}

export interface BatchTestConnectionResult {
  results: TestConnectionResult[]
}

export interface BatchDeleteResult {
  total: number
  success: number
  failed: number
}

export function importAssets(assets: CreateAssetParams[]): Promise<BatchImportResult> {
  return request.post('/assets/import', { assets })
}

export function testConnection(assetIds: number[]): Promise<BatchTestConnectionResult> {
  return request.post('/assets/test-connection', { asset_ids: assetIds })
}

export function batchDeleteAssets(ids: number[]): Promise<BatchDeleteResult> {
  return request.delete('/assets/batch', { data: { ids } })
}

// 资产初始化

export interface InitUserConfig {
  username: string
  shell?: string
  sudo?: boolean
}

export interface InitKeyConfig {
  ssh_key_id: number
  username: string
}

export interface AssetInitSteps {
  test_connection: boolean
  probe: boolean
  create_user?: InitUserConfig | null
  deploy_key?: InitKeyConfig | null
}

export interface InitStepResult {
  status: string
  message?: string
  latency?: number
}

export interface AssetInitResultItem {
  asset_id: number
  hostname: string
  ip: string
  status: string
  steps: {
    conn_test?: InitStepResult
    probe?: InitStepResult
    create_user?: InitStepResult
    deploy_key?: InitStepResult
  }
}

export interface AssetInitResponse {
  total: number
  success: number
  partial: number
  failed: number
  results: AssetInitResultItem[]
}

export interface AssetInitJobResponse {
  job_id: string
  total: number
}

export function initializeAssets(assetIds: number[], steps: AssetInitSteps): Promise<AssetInitJobResponse> {
  return request.post('/assets/init', { asset_ids: assetIds, steps })
}
