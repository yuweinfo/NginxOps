import request from './request'

export interface AuditLog {
  id: number
  userId: number
  username: string
  action: string
  module: string
  targetType: string
  targetId: number
  targetName: string
  detail: string
  ipAddress: string
  userAgent: string
  status: string
  errorMsg: string
  createdAt: string
}

export interface AuditListResponse {
  list: AuditLog[]
  total: number
  page: number
  size: number
}

export const auditApi = {
  list: (params: {
    page?: number
    size?: number
    module?: string
    action?: string
    userId?: number
    status?: string
    startTime?: string
    endTime?: string
  }) => {
    return request.get<AuditListResponse>('/audit', { params })
  },

  getById: (id: number) => {
    return request.get<AuditLog>(`/audit/${id}`)
  },

  getModules: () => {
    return request.get<string[]>('/audit/modules')
  },

  getActions: () => {
    return request.get<string[]>('/audit/actions')
  }
}
