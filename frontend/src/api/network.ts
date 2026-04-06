import request from './request'

export interface NetworkInterface {
  name: string
  ips: string[]
  mac: string
  flags: string[]
}

export interface DNSProviderInfo {
  id: number
  name: string
  providerType: string
}

export interface NetworkInfo {
  interfaces: NetworkInterface[]
  publicIp: string
  defaultDns: DNSProviderInfo | null
}

export interface CreateDNSRecordRequest {
  domain: string
  ip: string
  recordType?: string
  dnsProviderId?: number
}

export const networkApi = {
  // 获取网络信息（网卡、公网IP、默认DNS供应商）
  getInfo: () => request.get<NetworkInfo>('/network/info'),

  // 创建DNS解析记录
  createDNSRecord: (data: CreateDNSRecordRequest) =>
    request.post<{ domain: string; ip: string; type: string; provider: string }>('/network/dns-record', data),
}
