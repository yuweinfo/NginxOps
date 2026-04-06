import axios, { AxiosInstance, AxiosRequestConfig, AxiosError } from 'axios'

interface ApiResponse<T> {
  success: boolean
  message?: string
  data: T
}

// 扩展 AxiosError 类型
declare module 'axios' {
  interface AxiosError {
    userMessage?: string
  }
}

const TOKEN_KEY = 'nginxops_token'

const api: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

api.interceptors.response.use(
  (response) => response.data as any,
  (error) => {
    const status = error.response?.status
    const message = error.response?.data?.message || error.message || '请求失败'
    
    if (status === 401 || status === 403) {
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem('nginxops_user')
      window.location.href = '/login'
    }
    
    // 将后端错误消息附加到 error 对象
    error.userMessage = message
    return Promise.reject(error)
  }
)

// 重写 get/post 方法类型
const request = {
  get: <T>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> =>
    api.get(url, config) as any,

  post: <T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> =>
    api.post(url, data, config) as any,

  put: <T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> =>
    api.put(url, data, config) as any,

  delete: <T>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> =>
    api.delete(url, config) as any,
}

export default request
