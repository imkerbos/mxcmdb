import axios from 'axios'
import { message } from 'antd'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// 请求拦截：注入 Token
request.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截：统一处理错误
request.interceptors.response.use(
  (response) => {
    const { code, message: msg, data } = response.data
    if (code !== 0) {
      message.error(msg || '请求失败')
      if (code === 40100 || code === 40101) {
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        window.location.href = '/login'
      }
      return Promise.reject(new Error(msg))
    }
    return data
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      window.location.href = '/login'
    }
    const serverMsg = error.response?.data?.message
    const errMsg = serverMsg || error.message || '网络错误'
    message.error(errMsg)
    return Promise.reject(new Error(errMsg))
  },
)

export default request
