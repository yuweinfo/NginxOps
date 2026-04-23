import request from './request'

export interface SetupStatusResponse {
  configured: boolean
}

export interface SetupRequest {
  dbHost: string
  dbPort: number
  dbName: string
  dbUser: string
  dbPassword: string
  jwtSecret: string
  adminUsername: string
  adminEmail: string
  adminPassword: string
}

export interface DBTestRequest {
  host: string
  port: number
  name: string
  user: string
  password: string
}

export interface DBTestResponse {
  connected: boolean
  initialized: boolean
  message: string
}

export const setupApi = {
  getStatus: () => 
    request.get<SetupStatusResponse>('/setup/status'),

  initialize: (data: SetupRequest) =>
    request.post<{ success: boolean }>('/setup/init', data),

  testConnection: (data: DBTestRequest) =>
    request.post<DBTestResponse>('/setup/test-db', data),
}
