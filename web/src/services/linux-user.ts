import request from '../utils/request'
import type { PageResult } from './audit'

export interface LinuxUserItem {
  id: number
  username: string
  asset_id: number
  hostname: string
  ip: string
  uid: number
  gid: number
  home: string
  shell: string
  sudo: boolean
  status: string
}

export interface CreateLinuxUserParams {
  username: string
  asset_ids: number[]
  shell?: string
  sudo?: boolean
}

export function getLinuxUsers(params?: { username?: string; page?: number; page_size?: number }): Promise<PageResult<LinuxUserItem>> {
  return request.get('/linux-users', { params })
}

export function createLinuxUser(data: CreateLinuxUserParams): Promise<void> {
  return request.post('/linux-users', data)
}

export function deleteLinuxUser(username: string, assetIds: number[]): Promise<void> {
  return request.delete(`/linux-users/${username}`, { data: { asset_ids: assetIds } })
}
