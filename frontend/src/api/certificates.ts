import request from './request'

export interface Certificate {
  id: number
  domain: string
  issuer?: string
  issuedAt?: string
  expiresAt?: string
  status: string
  autoRenew: boolean
  certPath?: string
  keyPath?: string
  createdAt?: string
  updatedAt?: string
}

export interface CertificateImportData {
  certificate: string
  privateKey: string
  certificateChain?: string
  domain?: string
  remark?: string
}

export const certificatesApi = {
  list: () => request.get<Certificate[]>('/certificates'),

  listPage: (page = 1, size = 10, status = '') =>
    request.get<{ records: Certificate[], total: number }>('/certificates/page', {
      params: { page, size, status }
    }),

  getById: (id: number) =>
    request.get<Certificate>(`/certificates/${id}`),

  create: (data: Partial<Certificate>) =>
    request.post<Certificate>('/certificates', data),

  update: (id: number, data: Partial<Certificate>) =>
    request.put<Certificate>(`/certificates/${id}`, data),

  delete: (id: number) =>
    request.delete<void>(`/certificates/${id}`),

  request: (id: number, issuer: string, dnsProviderId?: number) =>
    request.post<void>(`/certificates/${id}/request`, { issuer, dnsProviderId }),

  renew: (id: number) =>
    request.post<Certificate>(`/certificates/${id}/renew`),

  toggleAutoRenew: (id: number, autoRenew: boolean) =>
    request.put<void>(`/certificates/${id}/auto-renew`, null, {
      params: { autoRenew }
    }),

  import: (data: CertificateImportData) =>
    request.post<Certificate>('/certificates/import', data),

  available: () =>
    request.get<Certificate[]>('/certificates/available'),
}
