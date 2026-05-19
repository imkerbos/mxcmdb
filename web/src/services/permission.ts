import request from '../utils/request'

export interface PermissionItem {
  id: number
  code: string
  name: string
  module: string
  description: string
}

export interface RolePermissions {
  role: string
  permissions: string[]
}

export function getAllPermissions(): Promise<PermissionItem[]> {
  return request.get('/permissions')
}

export function getMyPermissions(): Promise<RolePermissions> {
  return request.get('/permissions/mine')
}

export function getRolePermissions(role: string): Promise<RolePermissions> {
  return request.get(`/permissions/roles/${role}`)
}

export function setRolePermissions(role: string, permissions: string[]): Promise<void> {
  return request.put(`/permissions/roles/${role}`, { permissions })
}
