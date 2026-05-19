import request from '../utils/request'
import type { PageResult } from './audit'

export interface NotificationItem {
  id: number
  user_id: number
  title: string
  content: string
  type: string  // info / success / warning / error
  source: string
  source_id: number
  read: boolean
  created_at: string
}

export function getNotifications(params?: {
  read?: string
  page?: number
  page_size?: number
}): Promise<PageResult<NotificationItem>> {
  return request.get('/notifications', { params })
}

export function getUnreadCount(): Promise<{ count: number }> {
  return request.get('/notifications/unread-count')
}

export function markRead(id: number): Promise<void> {
  return request.put(`/notifications/${id}/read`)
}

export function markAllRead(): Promise<void> {
  return request.put('/notifications/read-all')
}

export function deleteNotification(id: number): Promise<void> {
  return request.delete(`/notifications/${id}`)
}
