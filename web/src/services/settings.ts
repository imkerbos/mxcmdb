import request from '../utils/request'

export interface SystemConfig {
  id: number
  key: string
  value: string
  category: string
  description: string
  type: string
}

export interface ConfigUpdateItem {
  key: string
  value: string
}

export function getConfigs(category?: string): Promise<SystemConfig[]> {
  return request.get('/settings/configs', { params: { category } })
}

export function updateConfig(key: string, value: string): Promise<void> {
  return request.put(`/settings/configs/${key}`, { value })
}

export function batchUpdateConfigs(configs: ConfigUpdateItem[]): Promise<void> {
  return request.put('/settings/configs', { configs })
}
