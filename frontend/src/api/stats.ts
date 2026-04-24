import request from './request'

export interface DashboardData {
  qps: number
  bandwidth: number
  pvToday: number
  activeSites: number
  hourlyTrend: Array<{ hour: string; requests: number }>
  statusDistribution: { '2xx': number; '3xx': number; '4xx': number; '5xx': number }
  ipLocations: Array<{ name: string; value: number[]; country?: string; region?: string }>
  ipRegionRank: Array<{ country: string; count: number; percent: number }>
  ipTopRank: Array<{ ip: string; region: string; requests: number }>
  hostRank: Array<{ name: string; count: number; percent: number }>
  refererRank: Array<{ name: string; count: number; percent: number }>
  pathRank: Array<{ name: string; count: number; percent: number }>
  resourceTypeRank: Array<{ name: string; count: number; percent: number }>
  statusRank: Array<{ name: string; count: number; percent: number }>
  browserRank: Array<{ name: string; count: number; percent: number }>
  deviceTypeRank: Array<{ name: string; count: number; percent: number }>
  osRank: Array<{ name: string; count: number; percent: number }>
  userAgentRank: Array<{ name: string; count: number; percent: number }>
}

export interface LogEntry {
  id: number
  time: string
  ip: string
  method: string
  url: string
  status: number
  size: number
  userAgent: string
  referer: string
  duration: number
}

export interface LogsPageData {
  records: LogEntry[]
  total: number
  size: number
  current: number
  pages: number
}

export const statsApi = {
  getDashboard: () => request.get<DashboardData>('/stats/dashboard'),

  queryLogs: (params: {
    start: string
    end: string
    ip?: string
    page?: number
    size?: number
  }) => request.get<LogsPageData>('/stats/logs', { params }),
}

export default statsApi
