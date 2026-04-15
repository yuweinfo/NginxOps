import request from './request'

// ==================== 类型定义 ====================

// 访问规则项
export interface AccessRuleItem {
  id: number
  ruleId: number
  itemType: 'ip' | 'geo'
  ipAddress: string
  countryCode: string
  action: 'allow' | 'block'  // IP 和 Geo 均使用：allow=允许, block=拒绝
  note: string
}

// 访问规则
export interface AccessRule {
  id: number
  name: string
  description: string
  enabled: boolean
  items: AccessRuleItem[]
}

// 访问规则摘要（列表使用）
export interface AccessRuleSummary {
  id: number
  name: string
  description: string
  enabled: boolean
  ipCount: number
  geoCount: number
  siteCount: number
  createdAt: string
  updatedAt: string
}

// 创建/更新规则的 DTO
export interface AccessRuleDto {
  id?: number
  name: string
  description: string
  enabled: boolean
  items: AccessRuleItem[]
}

// ==================== API ====================

export const accessRuleApi = {
  // ========== 规则 CRUD ==========
  list: () =>
    request.get<AccessRuleSummary[]>('/access-rules'),

  getById: (id: number) =>
    request.get<AccessRule>(`/access-rules/${id}`),

  create: (data: AccessRuleDto) =>
    request.post<AccessRule>('/access-rules', data),

  update: (id: number, data: AccessRuleDto) =>
    request.put<AccessRule>(`/access-rules/${id}`, data),

  delete: (id: number) =>
    request.delete<void>(`/access-rules/${id}`),

  toggle: (id: number, enabled: boolean) =>
    request.put<void>(`/access-rules/${id}/toggle`, null, { params: { enabled } }),

  // ========== 站点-规则关联 ==========
  getSiteRuleIds: (siteId: number) =>
    request.get<number[]>(`/access-rules/sites/${siteId}/rules`),

  setSiteRules: (siteId: number, ruleIds: number[]) =>
    request.put<{ message: string }>(`/access-rules/sites/${siteId}/rules`, { ruleIds }),

  // ========== 配置同步 ==========
  syncConfig: () =>
    request.post<{ message: string }>('/access-rules/sync'),
}
