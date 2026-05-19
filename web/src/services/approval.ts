import request from '../utils/request'
import type { PageResult } from './audit'

export interface ApprovalItem {
  id: number
  title: string
  description: string
  type: string
  payload: string
  status: string
  requested_by: number
  requester_name: string
  reviewed_by: number | null
  reviewer_name: string
  reviewed_at: string
  review_note: string
  created_at: string
}

export interface CreateApprovalParams {
  title: string
  description?: string
  type: string
  payload: string
}

export interface ReviewApprovalParams {
  action: 'approve' | 'reject'
  note?: string
}

export function getApprovals(params?: {
  status?: string
  type?: string
  page?: number
  page_size?: number
}): Promise<PageResult<ApprovalItem>> {
  return request.get('/approvals', { params })
}

export function getMyApprovals(params?: {
  page?: number
  page_size?: number
}): Promise<PageResult<ApprovalItem>> {
  return request.get('/approvals/mine', { params })
}

export function getApproval(id: number): Promise<ApprovalItem> {
  return request.get(`/approvals/${id}`)
}

export function createApproval(data: CreateApprovalParams): Promise<ApprovalItem> {
  return request.post('/approvals', data)
}

export function reviewApproval(id: number, data: ReviewApprovalParams): Promise<ApprovalItem> {
  return request.put(`/approvals/${id}/review`, data)
}

export function cancelApproval(id: number): Promise<void> {
  return request.put(`/approvals/${id}/cancel`)
}

export function getPendingCount(): Promise<{ count: number }> {
  return request.get('/approvals/pending-count')
}
