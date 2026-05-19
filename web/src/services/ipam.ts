import request from '../utils/request'
import type { PageResult } from './audit'

export interface SubnetItem {
  id: number
  name: string
  cidr: string
  gateway: string
  vlan: number
  total_ips: number
  used_ips: number
  usage_percent: number
  comment: string
}

export interface IPAddressItem {
  id: number
  subnet_id: number
  address: string
  status: string
  asset_id: number | null
  hostname: string
  comment: string
}

export interface CreateSubnetParams {
  name: string
  cidr: string
  gateway?: string
  vlan?: number
  comment?: string
}

export function getSubnets(params?: { page?: number; page_size?: number }): Promise<PageResult<SubnetItem>> {
  return request.get('/ipam/subnets', { params })
}

export function createSubnet(data: CreateSubnetParams): Promise<SubnetItem> {
  return request.post('/ipam/subnets', data)
}

export function updateSubnet(id: number, data: Partial<CreateSubnetParams>): Promise<void> {
  return request.put(`/ipam/subnets/${id}`, data)
}

export function deleteSubnet(id: number): Promise<void> {
  return request.delete(`/ipam/subnets/${id}`)
}

export function getSubnetIPs(id: number, params?: { status?: string; page?: number; page_size?: number }): Promise<PageResult<IPAddressItem>> {
  return request.get(`/ipam/subnets/${id}/ips`, { params })
}

export function allocateIP(ipId: number, data: { asset_id?: number; hostname?: string; comment?: string }): Promise<void> {
  return request.post(`/ipam/ips/${ipId}/allocate`, data)
}

export function releaseIP(ipId: number): Promise<void> {
  return request.post(`/ipam/ips/${ipId}/release`)
}
