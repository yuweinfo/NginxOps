import request from './request'

export interface DnsProvider {
  id: number
  name: string
  providerType: 'aliyun' | 'tencent' | 'cloudflare'
  accessKeyId: string
  accessKeySecret?: string
  isDefault: boolean
  createdAt?: string
  updatedAt?: string
}

export const dnsProvidersApi = {
  list: () => request.get<DnsProvider[]>('/dns-providers'),

  getById: (id: number) =>
    request.get<DnsProvider>(`/dns-providers/${id}`),

  create: (data: Partial<DnsProvider>) =>
    request.post<DnsProvider>('/dns-providers', data),

  update: (id: number, data: Partial<DnsProvider>) =>
    request.put<DnsProvider>(`/dns-providers/${id}`, data),

  delete: (id: number) =>
    request.delete<void>(`/dns-providers/${id}`),

  setDefault: (id: number) =>
    request.put<void>(`/dns-providers/${id}/default`),
}
