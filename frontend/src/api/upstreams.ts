import request from './request'

export interface UpstreamServer {
  id?: number
  host: string
  port: number
  weight?: number
  maxFails?: number
  failTimeout?: number
  status?: 'up' | 'down' | 'checking'
  backup?: boolean
}

export interface Upstream {
  id?: number
  name: string
  lbMode: 'round_robin' | 'weight' | 'ip_hash' | 'least_conn'
  healthCheck: boolean
  checkInterval?: number
  checkPath?: string
  checkTimeout?: number
  servers: UpstreamServer[]
  createdAt?: string
  updatedAt?: string
}

export interface HealthCheckResult {
  upstreamId: number
  upstreamName: string
  serverHost: string
  serverPort: number
  healthy: boolean
  responseTime: number
  error?: string
  checkedAt: string
}

export const upstreamsApi = {
  list: () => request.get<Upstream[]>('/upstreams'),

  listPage: (page = 1, size = 10) =>
    request.get<{ records: any[], total: number }>('/upstreams/page', {
      params: { page, size }
    }),

  getById: (id: number) => request.get<Upstream>(`/upstreams/${id}`),

  create: (data: Upstream) => request.post<Upstream>('/upstreams', data),

  update: (id: number, data: Upstream) => request.put<Upstream>(`/upstreams/${id}`, data),

  delete: (id: number) => request.delete<void>(`/upstreams/${id}`),

  getConfig: (id: number) => request.get<string>(`/upstreams/${id}/config`),

  healthCheck: (id: number) => request.post<HealthCheckResult[]>(`/upstreams/${id}/health-check`),
}
