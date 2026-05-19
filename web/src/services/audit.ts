import request from '../utils/request'

export interface AuditLog {
  id: number
  created_at: string
  user_id: number
  username: string
  module: string
  action: string
  resource: string
  resource_id: number
  client_ip: string
  method: string
  path: string
  request_body: string
  response_code: number
  duration: number
  detail: string
}

export interface AuditLogListParams {
  module?: string
  action?: string
  username?: string
  start_date?: string
  end_date?: string
  page?: number
  page_size?: number
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export function getAuditLogs(params?: AuditLogListParams): Promise<PageResult<AuditLog>> {
  return request.get('/audit/logs', { params })
}

export function getAuditLogDetail(id: number): Promise<AuditLog> {
  return request.get(`/audit/logs/${id}`)
}
