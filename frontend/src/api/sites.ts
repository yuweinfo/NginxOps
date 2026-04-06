import request from './request'

// Location 配置
export interface LocationConfig {
  path: string
  proxyPass: string
  proxyHeaders: boolean
  websocket: boolean
}

// Upstream 服务器配置 (站点专用)
export interface SiteUpstreamServer {
  address: string
  weight: number
  backup: boolean
}

// 站点实体
export interface Site {
  id: number
  fileName: string
  domain: string
  port: number
  siteType: 'static' | 'proxy' | 'loadbalance'
  rootDir: string
  locations: LocationConfig[]
  upstreamServers: SiteUpstreamServer[]
  sslEnabled: boolean
  certId: number | null
  forceHttps: boolean
  gzip: boolean
  cache: boolean
  enabled: boolean
  config: string
  createdAt?: string
  updatedAt?: string
}

// 站点 DTO
export interface SiteDto {
  id?: number
  fileName?: string
  domain?: string
  port?: number
  siteType?: 'static' | 'proxy' | 'loadbalance'
  rootDir?: string
  locations?: string  // JSON string
  upstreamServers?: string  // JSON string
  sslEnabled?: boolean
  certId?: number | null
  forceHttps?: boolean
  gzip?: boolean
  cache?: boolean
  enabled?: boolean
  config?: string
}

export const sitesApi = {
  list: () => request.get<Site[]>('/sites'),

  listPage: (page = 1, size = 10, keyword = '') =>
    request.get<{ records: Site[], total: number }>('/sites/page', {
      params: { page, size, keyword }
    }),

  getById: (id: number) => request.get<Site>(`/sites/${id}`),

  create: (data: SiteDto) => request.post<Site>('/sites', data),

  update: (id: number, data: SiteDto) => request.put<Site>(`/sites/${id}`, data),

  delete: (id: number) => request.delete<void>(`/sites/${id}`),

  toggleEnabled: (id: number, enabled: boolean) =>
    request.put<Site>(`/sites/${id}/toggle`, null, { params: { enabled } }),

  getConfig: (id: number) => request.get<string>(`/sites/${id}/config`),

  sync: () => request.post<void>('/sites/sync'),
}
