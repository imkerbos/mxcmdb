import request from '../utils/request'

export interface LoginParams {
  username: string
  password: string
  captcha_id: string
  captcha: string
  mfa_code?: string
}

export interface LoginResult {
  access_token: string
  refresh_token: string
  expires_in: number
  require_mfa: boolean
  require_mfa_setup: boolean
}

export interface CaptchaResult {
  captcha_id: string
  captcha_image: string
}

export interface UserInfo {
  id: number
  username: string
  nickname: string
  email: string
  phone: string
  role: string
  mfa_enabled: boolean
}

export interface MFASetupResult {
  secret: string
  url: string
}

export function getCaptcha(): Promise<CaptchaResult> {
  return request.get('/auth/captcha')
}

export function login(params: LoginParams): Promise<LoginResult> {
  return request.post('/auth/login', params)
}

export function getUserInfo(): Promise<UserInfo> {
  return request.get('/user/info')
}

export function setupMFA(): Promise<MFASetupResult> {
  return request.get('/user/mfa/setup')
}

export function bindMFA(secret: string, code: string): Promise<void> {
  return request.post(`/user/mfa/bind?secret=${secret}`, { code })
}
