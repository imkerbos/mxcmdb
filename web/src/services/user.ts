import request from '../utils/request'
import type { PageResult } from './audit'

export interface UserItem {
  id: number
  username: string
  nickname: string
  email: string
  phone: string
  avatar: string
  role: string
  status: number
  mfa_enabled: boolean
  last_login_at: string
  last_login_ip: string
  created_at: string
}

export interface UserListParams {
  keyword?: string
  role?: string
  status?: number
  page?: number
  page_size?: number
}

export interface CreateUserParams {
  username: string
  password: string
  nickname?: string
  email?: string
  phone?: string
  role: string
}

export interface UpdateUserParams {
  nickname?: string
  email?: string
  phone?: string
  role?: string
}

export function getUsers(params?: UserListParams): Promise<PageResult<UserItem>> {
  return request.get('/users', { params })
}

export function createUser(data: CreateUserParams): Promise<void> {
  return request.post('/users', data)
}

export function updateUser(id: number, data: UpdateUserParams): Promise<void> {
  return request.put(`/users/${id}`, data)
}

export function deleteUser(id: number): Promise<void> {
  return request.delete(`/users/${id}`)
}

export function resetPassword(id: number, password: string): Promise<void> {
  return request.put(`/users/${id}/password`, { password })
}

export function toggleUserStatus(id: number, status: number): Promise<void> {
  return request.put(`/users/${id}/status`, { status })
}

export function changePassword(oldPassword: string, newPassword: string): Promise<void> {
  return request.put('/user/password', { old_password: oldPassword, new_password: newPassword })
}
