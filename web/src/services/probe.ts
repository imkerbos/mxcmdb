import request from '../utils/request'

export interface ProbeResult {
  id: number
  asset_id: number
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
  collected_at: string
  status: string
  error_message: string
}

export function executeProbe(assetIds: number[]): Promise<void> {
  return request.post('/probe/execute', { asset_ids: assetIds })
}

export function getLatestProbe(assetId: number): Promise<ProbeResult> {
  return request.get(`/assets/${assetId}/probe`)
}

export function getProbeHistory(assetId: number): Promise<ProbeResult[]> {
  return request.get(`/assets/${assetId}/probe/history`)
}
