import request from './request'

export interface UserInfo {
  id: number
  username: string
  email?: string
  role: string
  createdAt?: string
}

export const userApi = {
  // 获取当前用户信息
  getCurrentUser: () => {
    return request.get<UserInfo>('/users/me')
  },

  // 验证密码
  verifyPassword: (password: string) => {
    return request.post<{ valid: boolean }>('/users/verify-password', { password })
  },

  // 更新个人信息
  updateProfile: (data: { email?: string; password?: string; oldPassword?: string }) => {
    return request.put<UserInfo>('/users/me', data)
  }
}
