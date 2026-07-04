import { apiClient } from '../client'

export type LicenseCodeStatus = 'unused' | 'active' | 'disabled' | 'expired' | 'revoked' | 'refunded'

export interface LicenseCode {
  id: number
  codeId: string
  code: string
  licenseId: string
  product: string
  productBatch: string
  features: string[]
  status: LicenseCodeStatus
  usbFingerprint: string
  activatedAt: string
  lastVerifiedAt: string
  expiresAt: string
  revokedAt: string
  revokedReason: string
  createdAt: string
  updatedAt: string
}

export interface CreateLicenseCodesRequest {
  count: number
  product?: string
  productBatch?: string
  features?: string[]
  prefix?: string
  expiresAt?: string
}

export interface UpdateCodeFeaturesRequest {
  features: string[]
}

export async function listCodes(): Promise<LicenseCode[]> {
  const { data } = await apiClient.get<{ codes: LicenseCode[] }>('/admin/license/codes')
  return data.codes
}

export async function createCodes(payload: CreateLicenseCodesRequest): Promise<LicenseCode[]> {
  const { data } = await apiClient.post<{ codes: LicenseCode[] }>('/admin/license/codes', payload)
  return data.codes
}

export async function disableCode(codeId: string): Promise<LicenseCode> {
  const { data } = await apiClient.post<LicenseCode>(`/admin/license/codes/${codeId}/disable`)
  return data
}

export async function enableCode(codeId: string): Promise<LicenseCode> {
  const { data } = await apiClient.post<LicenseCode>(`/admin/license/codes/${codeId}/enable`)
  return data
}

export async function refundCode(codeId: string): Promise<LicenseCode> {
  const { data } = await apiClient.post<LicenseCode>(`/admin/license/codes/${codeId}/refund`)
  return data
}

export async function revokeLicense(licenseId: string): Promise<LicenseCode> {
  const { data } = await apiClient.post<LicenseCode>(`/admin/license/licenses/${licenseId}/revoke`)
  return data
}

export async function updateCodeFeatures(codeId: string, payload: UpdateCodeFeaturesRequest): Promise<LicenseCode> {
  const { data } = await apiClient.put<LicenseCode>(`/admin/license/codes/${codeId}/features`, payload)
  return data
}

export const licenseAPI = {
  listCodes,
  createCodes,
  disableCode,
  enableCode,
  refundCode,
  revokeLicense,
  updateCodeFeatures,
}

export default licenseAPI
