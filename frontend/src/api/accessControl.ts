import request from './request'

// ==================== 类型定义 ====================

// 全局设置
export interface AccessControlSettings {
  geoEnabled: boolean
  ipBlacklistEnabled: boolean
  defaultAction: 'allow' | 'block'
}

// IP 黑名单项
export interface IPBlacklistItem {
  id: number
  ipAddress: string
  note: string
  enabled: boolean
  createdAt: string
  updatedAt: string
}

// Geo 规则项
export interface GeoRuleItem {
  id: number
  countryCode: string
  action: 'allow' | 'block'
  note: string
  enabled: boolean
  createdAt: string
  updatedAt: string
}

// 站点 IP 黑名单项
export interface SiteIPBlacklistItem {
  id: number
  siteId: number
  ipAddress: string
  note: string
  enabled: boolean
  createdAt: string
  updatedAt: string
}

// 站点 Geo 规则项
export interface SiteGeoRuleItem {
  id: number
  siteId: number
  countryCode: string
  action: 'allow' | 'block'
  note: string
  enabled: boolean
  createdAt: string
  updatedAt: string
}

// ==================== API ====================

export const accessControlApi = {
  // ========== 全局设置 ==========
  getSettings: () => 
    request.get<AccessControlSettings>('/access-control/settings'),
  
  updateSettings: (data: AccessControlSettings) => 
    request.put<{ message: string }>('/access-control/settings', data),

  // ========== 全局 IP 黑名单 ==========
  listIPBlacklist: () => 
    request.get<IPBlacklistItem[]>('/access-control/ip-blacklist'),
  
  createIPBlacklist: (data: { ipAddress: string; note?: string; enabled?: boolean }) => 
    request.post<IPBlacklistItem>('/access-control/ip-blacklist', data),
  
  updateIPBlacklist: (id: number, data: { ipAddress?: string; note?: string; enabled?: boolean }) => 
    request.put<IPBlacklistItem>(`/access-control/ip-blacklist/${id}`, data),
  
  deleteIPBlacklist: (id: number) => 
    request.delete<void>(`/access-control/ip-blacklist/${id}`),
  
  toggleIPBlacklist: (id: number, enabled: boolean) => 
    request.put<void>(`/access-control/ip-blacklist/${id}/toggle`, null, { params: { enabled } }),

  // ========== 全局 Geo 规则 ==========
  listGeoRules: () => 
    request.get<GeoRuleItem[]>('/access-control/geo-rules'),
  
  createGeoRule: (data: { countryCode: string; action: 'allow' | 'block'; note?: string; enabled?: boolean }) => 
    request.post<GeoRuleItem>('/access-control/geo-rules', data),
  
  updateGeoRule: (id: number, data: { countryCode?: string; action?: 'allow' | 'block'; note?: string; enabled?: boolean }) => 
    request.put<GeoRuleItem>(`/access-control/geo-rules/${id}`, data),
  
  deleteGeoRule: (id: number) => 
    request.delete<void>(`/access-control/geo-rules/${id}`),
  
  toggleGeoRule: (id: number, enabled: boolean) => 
    request.put<void>(`/access-control/geo-rules/${id}/toggle`, null, { params: { enabled } }),

  // ========== 站点专属 IP 黑名单 ==========
  listSiteIPBlacklist: (siteId: number) => 
    request.get<SiteIPBlacklistItem[]>(`/access-control/sites/${siteId}/ip-blacklist`),
  
  createSiteIPBlacklist: (siteId: number, data: { ipAddress: string; note?: string; enabled?: boolean }) => 
    request.post<SiteIPBlacklistItem>(`/access-control/sites/${siteId}/ip-blacklist`, data),
  
  deleteSiteIPBlacklist: (id: number) => 
    request.delete<void>(`/access-control/site-ip-blacklist/${id}`),

  // ========== 站点专属 Geo 规则 ==========
  listSiteGeoRules: (siteId: number) => 
    request.get<SiteGeoRuleItem[]>(`/access-control/sites/${siteId}/geo-rules`),
  
  createSiteGeoRule: (siteId: number, data: { countryCode: string; action: 'allow' | 'block'; note?: string; enabled?: boolean }) => 
    request.post<SiteGeoRuleItem>(`/access-control/sites/${siteId}/geo-rules`, data),
  
  deleteSiteGeoRule: (id: number) => 
    request.delete<void>(`/access-control/site-geo-rules/${id}`),

  // ========== 配置同步 ==========
  syncConfig: () => 
    request.post<{ message: string }>('/access-control/sync'),
}

// 国家代码映射表
export const COUNTRY_CODES: Record<string, { name: string; flag: string }> = {
  CN: { name: '中国', flag: '🇨🇳' },
  US: { name: '美国', flag: '🇺🇸' },
  RU: { name: '俄罗斯', flag: '🇷🇺' },
  KP: { name: '朝鲜', flag: '🇰🇵' },
  KR: { name: '韩国', flag: '🇰🇷' },
  JP: { name: '日本', flag: '🇯🇵' },
  HK: { name: '香港', flag: '🇭🇰' },
  TW: { name: '台湾', flag: '🇹🇼' },
  SG: { name: '新加坡', flag: '🇸🇬' },
  DE: { name: '德国', flag: '🇩🇪' },
  FR: { name: '法国', flag: '🇫🇷' },
  GB: { name: '英国', flag: '🇬🇧' },
  AU: { name: '澳大利亚', flag: '🇦🇺' },
  CA: { name: '加拿大', flag: '🇨🇦' },
  IN: { name: '印度', flag: '🇮🇳' },
  BR: { name: '巴西', flag: '🇧🇷' },
  IT: { name: '意大利', flag: '🇮🇹' },
  ES: { name: '西班牙', flag: '🇪🇸' },
  NL: { name: '荷兰', flag: '🇳🇱' },
  UA: { name: '乌克兰', flag: '🇺🇦' },
  IR: { name: '伊朗', flag: '🇮🇷' },
  NG: { name: '尼日利亚', flag: '🇳🇬' },
  ZA: { name: '南非', flag: '🇿🇦' },
  TR: { name: '土耳其', flag: '🇹🇷' },
  SA: { name: '沙特阿拉伯', flag: '🇸🇦' },
  MX: { name: '墨西哥', flag: '🇲🇽' },
  ID: { name: '印度尼西亚', flag: '🇮🇩' },
  TH: { name: '泰国', flag: '🇹🇭' },
  VN: { name: '越南', flag: '🇻🇳' },
  MY: { name: '马来西亚', flag: '🇲🇾' },
  PH: { name: '菲律宾', flag: '🇵🇭' },
  AE: { name: '阿联酋', flag: '🇦🇪' },
  IL: { name: '以色列', flag: '🇮🇱' },
  PL: { name: '波兰', flag: '🇵🇱' },
  SE: { name: '瑞典', flag: '🇸🇪' },
  NO: { name: '挪威', flag: '🇳🇴' },
  CH: { name: '瑞士', flag: '🇨🇭' },
  BE: { name: '比利时', flag: '🇧🇪' },
  AT: { name: '奥地利', flag: '🇦🇹' },
  CZ: { name: '捷克', flag: '🇨🇿' },
  PT: { name: '葡萄牙', flag: '🇵🇹' },
  DK: { name: '丹麦', flag: '🇩🇰' },
  FI: { name: '芬兰', flag: '🇫🇮' },
  IE: { name: '爱尔兰', flag: '🇮🇪' },
  NZ: { name: '新西兰', flag: '🇳🇿' },
  AR: { name: '阿根廷', flag: '🇦🇷' },
  CL: { name: '智利', flag: '🇨🇱' },
  CO: { name: '哥伦比亚', flag: '🇨🇴' },
  EG: { name: '埃及', flag: '🇪🇬' },
  PK: { name: '巴基斯坦', flag: '🇵🇰' },
  BD: { name: '孟加拉', flag: '🇧🇩' },
}

export const getCountryInfo = (code: string) => {
  return COUNTRY_CODES[code] || { name: code, flag: '🏳️' }
}
