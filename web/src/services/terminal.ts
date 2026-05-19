import request from '../utils/request'
import type { PageResult } from './audit'

export interface TerminalSession {
  id: number
  asset_id: number
  hostname: string
  ip: string
  user_id: number
  username: string
  status: string
  client_ip: string
  started_at: string
  finished_at: string
}

export interface ActiveSession {
  session_id: number
  asset_id: number
  hostname: string
  ip: string
  user_id: number
  username: string
  client_ip: string
  started_at: string
  duration: string
}

export function getTerminalSessions(params?: { page?: number; page_size?: number; status?: string }): Promise<PageResult<TerminalSession>> {
  return request.get('/terminal/sessions', { params })
}

export function getActiveSessions(): Promise<ActiveSession[]> {
  return request.get('/terminal/sessions/active')
}

export function killSession(id: number): Promise<void> {
  return request.delete(`/terminal/sessions/${id}`)
}

export function getSessionRecording(id: number): Promise<string> {
  return request.get(`/terminal/sessions/${id}/recording`)
}

export function connectTerminalWS(assetId: number): WebSocket | null {
  const token = localStorage.getItem('access_token')
  if (!token) return null

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const url = `${protocol}//${host}/ws/terminal/${assetId}?token=${token}`

  return new WebSocket(url)
}
