import request from '../utils/request'
import type { PageResult } from './audit'

export interface CloudAccount {
  id: number
  name: string
  provider: string
  access_key_id: string
  access_key_secret: string
  region: string
  status: number
  last_sync_at: string
  last_sync_status: string
  created_at: string
}

export interface CreateCloudAccountParams {
  name: string
  provider: string
  access_key_id: string
  access_key_secret: string
  region?: string
}

export interface UpdateCloudAccountParams {
  name?: string
  access_key_id?: string
  access_key_secret?: string
  region?: string
}

export function getCloudAccounts(params?: Record<string, unknown>): Promise<PageResult<CloudAccount>> {
  return request.get('/cloud-accounts', { params })
}

export function getCloudAccount(id: number): Promise<CloudAccount> {
  return request.get(`/cloud-accounts/${id}`)
}

export function createCloudAccount(data: CreateCloudAccountParams): Promise<void> {
  return request.post('/cloud-accounts', data)
}

export function updateCloudAccount(id: number, data: UpdateCloudAccountParams): Promise<void> {
  return request.put(`/cloud-accounts/${id}`, data)
}

export function deleteCloudAccount(id: number): Promise<void> {
  return request.delete(`/cloud-accounts/${id}`)
}
