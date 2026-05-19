import request from '../utils/request'

export interface DistributionItem {
  label: string
  value: number
}

export interface DashboardStats {
  total_assets: number
  cloud_assets: number
  idc_assets: number
  online_assets: number
  offline_assets: number
  probed_assets: number
  total_projects: number
  env_distribution: DistributionItem[]
  type_distribution: DistributionItem[]
}

export interface RecentActivity {
  id: number
  username: string
  module: string
  action: string
  path: string
  client_ip: string
  created_at: string
}

export function getDashboardStats(): Promise<DashboardStats> {
  return request.get('/dashboard/stats')
}

export function getRecentActivities(): Promise<RecentActivity[]> {
  return request.get('/dashboard/activities')
}

export interface QuickActionsData {
  actions: string[]
}

export function getQuickActions(): Promise<QuickActionsData> {
  return request.get('/dashboard/quick-actions')
}

export function updateQuickActions(actions: string[]): Promise<void> {
  return request.put('/dashboard/quick-actions', { actions })
}
