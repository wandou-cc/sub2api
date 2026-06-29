import { DEFAULT_STREAM_PARTIAL_IMAGES, type ApiProfile, type AppSettings } from '../types'
import { DEFAULT_IMAGES_MODEL, DEFAULT_SETTINGS, normalizeSettings } from './apiProfiles'

export const CODEINGFORCE_PROVIDER_NAME = 'CodingForce'
export const CODEINGFORCE_API_URL = 'https://codeingforce.com/v1'
export const CODEINGFORCE_DASHBOARD_URL = 'https://codeingforce.com/dashboard'

export interface CodeingForceApiKey {
  id: number
  key: string
  name: string
}

interface CodeingForceApiKeysResponse {
  code: number
  message: string
  data: {
    items: CodeingForceApiKey[]
  }
}

// 读取当前登录用户自己的可用 API Key，后端按登录态 user_id 过滤。
export async function fetchCurrentUserCodeingForceApiKeys(signal?: AbortSignal): Promise<CodeingForceApiKey[]> {
  const token = window.localStorage.getItem('auth_token')
  if (!token) throw new Error('当前未登录主系统，无法读取已配置 Key')

  const params = new URLSearchParams({
    page: '1',
    page_size: '1000',
    status: 'active',
    sort_by: 'created_at',
    sort_order: 'desc',
    _: String(Date.now()),
  })
  const response = await fetch(`/api/v1/keys?${params.toString()}`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    cache: 'no-store',
    signal,
  })
  const payload = await response.json() as CodeingForceApiKeysResponse
  if (!response.ok || payload.code !== 0) throw new Error(payload.message)

  return payload.data.items
}

// 将生图配置固定为 CodingForce，并写入当前用户最新的 API Key。
export function applyCurrentUserCodeingForceApiKey(settings: AppSettings, apiKey: CodeingForceApiKey | null, apiProxyAvailable: boolean): AppSettings {
  const normalized = normalizeSettings(settings)
  return normalizeSettings({
    ...normalized,
    profiles: normalized.profiles.map((profile) => applyCodeingForceProfile(profile, apiKey?.key ?? '', apiProxyAvailable)),
  })
}

// 统一单个 API 配置的 CodingForce 服务商字段。
export function applyCodeingForceProfile(profile: ApiProfile, apiKey: string, apiProxyAvailable: boolean): ApiProfile {
  return {
    ...profile,
    name: CODEINGFORCE_PROVIDER_NAME,
    provider: 'openai',
    baseUrl: CODEINGFORCE_API_URL,
    apiKey,
    model: profile.model.trim() || DEFAULT_IMAGES_MODEL,
    timeout: Number(profile.timeout) || DEFAULT_SETTINGS.timeout,
    apiMode: 'images',
    apiProxy: apiProxyAvailable,
    codexCli: false,
    streamImages: false,
    streamPartialImages: DEFAULT_STREAM_PARTIAL_IMAGES,
  }
}
