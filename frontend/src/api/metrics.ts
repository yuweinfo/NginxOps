import request from './request'

export interface MetricsOverview {
  totalRequests: number
  totalBytes: number
  peakQPS: number
  avgRT: number
}

export interface TrendPoint {
  time: string
  requests?: number
  bytes?: number
  p50?: number
  p90?: number
  p99?: number
  count?: number
  total?: number
  errors?: number
  errorRate?: number
}

export interface DistributionItem {
  name: string
  count: number
  percent: number
}

export interface ErrorPathItem {
  path: string
  count: number
  percent: number
}

export interface ClientAnalysis {
  deviceTypeRank: DistributionItem[]
  browserRank: DistributionItem[]
  osRank: DistributionItem[]
  userAgentRank: DistributionItem[]
}

export const metricsApi = {
  getOverview: (params: { start: string; end: string }) =>
    request.get<MetricsOverview>('/metrics/overview', { params }),

  getTrafficTrend: (params: { start: string; end: string; granularity?: string }) =>
    request.get<TrendPoint[]>('/metrics/traffic', { params }),

  getResponseTrend: (params: { start: string; end: string; granularity?: string }) =>
    request.get<TrendPoint[]>('/metrics/response', { params }),

  getSlowRequestTrend: (params: { start: string; end: string; granularity?: string }) =>
    request.get<TrendPoint[]>('/metrics/slow-requests', { params }),

  getMethodDistribution: (params: { start: string; end: string }) =>
    request.get<DistributionItem[]>('/metrics/method-distribution', { params }),

  getStatusDistribution: (params: { start: string; end: string }) =>
    request.get<DistributionItem[]>('/metrics/status-distribution', { params }),

  getErrorRateTrend: (params: { start: string; end: string; granularity?: string }) =>
    request.get<TrendPoint[]>('/metrics/error-rate', { params }),

  getErrorPaths: (params: { start: string; end: string }) =>
    request.get<ErrorPathItem[]>('/metrics/error-paths', { params }),

  getClientAnalysis: (params: { start: string; end: string }) =>
    request.get<ClientAnalysis>('/metrics/client', { params }),
}
