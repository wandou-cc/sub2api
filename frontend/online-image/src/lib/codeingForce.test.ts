import { afterEach, describe, expect, it, vi } from 'vitest'
import { DEFAULT_SETTINGS, normalizeSettings } from './apiProfiles'
import {
  applyCurrentUserCodeingForceApiKey,
  CODEINGFORCE_API_URL,
  CODEINGFORCE_PROVIDER_NAME,
  fetchCurrentUserCodeingForceApiKeys,
} from './codeingForce'

afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

describe('fetchCurrentUserCodeingForceApiKeys', () => {
  it('loads active keys from the current user endpoint with the auth token', async () => {
    vi.stubGlobal('window', {
      localStorage: {
        getItem: vi.fn(() => 'user-token'),
      },
    })
    vi.spyOn(Date, 'now').mockReturnValue(123)
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => ({
      ok: true,
      json: async () => ({
        code: 0,
        message: 'success',
        data: {
          items: [
            { id: 2, key: 'new-key', name: 'new' },
            { id: 1, key: 'old-key', name: 'old' },
          ],
        },
      }),
    }))
    vi.stubGlobal('fetch', fetchMock)

    const keys = await fetchCurrentUserCodeingForceApiKeys()

    expect(keys[0]).toMatchObject({ id: 2, key: 'new-key' })
    expect(fetchMock).toHaveBeenCalledTimes(1)
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/v1/keys?page=1&page_size=1000&status=active&sort_by=created_at&sort_order=desc&_=123')
    expect(init).toMatchObject({
      headers: {
        Authorization: 'Bearer user-token',
      },
      cache: 'no-store',
    })
  })
})

describe('applyCurrentUserCodeingForceApiKey', () => {
  it('replaces stale image API keys with the latest current-user key', () => {
    const settings = normalizeSettings({
      ...DEFAULT_SETTINGS,
      apiKey: 'old-key',
    })

    const synced = applyCurrentUserCodeingForceApiKey(settings, { id: 2, key: 'new-key', name: 'new' }, true)

    expect(synced.apiKey).toBe('new-key')
    expect(synced.profiles[0]).toMatchObject({
      name: CODEINGFORCE_PROVIDER_NAME,
      provider: 'openai',
      baseUrl: CODEINGFORCE_API_URL,
      apiKey: 'new-key',
      apiMode: 'images',
      apiProxy: true,
      codexCli: false,
      streamImages: false,
    })
  })

  it('clears stale image API keys when the current user has no active key', () => {
    const settings = normalizeSettings({
      ...DEFAULT_SETTINGS,
      apiKey: 'deleted-key',
    })

    const synced = applyCurrentUserCodeingForceApiKey(settings, null, true)

    expect(synced.apiKey).toBe('')
    expect(synced.profiles[0].apiKey).toBe('')
  })
})
