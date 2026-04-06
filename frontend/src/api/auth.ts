import request from './request'

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  username: string
  role: string
}

export const authApi = {
  login: (data: LoginRequest) =>
    request.post<LoginResponse>('/auth/login', data),

  logout: () =>
    request.post<void>('/auth/logout'),

  getCurrentUser: () =>
    request.get<LoginResponse>('/auth/me'),
}

export default authApi
