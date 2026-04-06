import request from './request'

export interface NginxStatus {
  running: boolean
  version: string
  uptime: string
  workers: number
  activeConnections: number
  requestsPerSecond: number
  memoryUsage: string
}

export interface ConfigHistory {
  id: number
  configName: string
  operator: string
  remark: string
  createdAt: string
  testResult: boolean
}

export const nginxApi = {
  getStatus: () => request.get<NginxStatus>('/nginx/status'),

  start: () => request.post<string>('/nginx/start'),

  stop: () => request.post<string>('/nginx/stop'),

  restart: () => request.post<string>('/nginx/restart'),

  reload: () => request.post<string>('/nginx/reload'),

  testConfig: () => request.post<{ success: boolean, message: string }>('/nginx/test'),

  getHistory: (configName?: string, limit = 20) =>
    request.get<ConfigHistory[]>('/nginx/history', {
      params: { configName, limit }
    }),
}
