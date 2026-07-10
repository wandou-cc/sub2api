import { afterEach, describe, expect, it, vi } from 'vitest'
import { buildApiUrl, isApiProxyAvailable, shouldUseApiProxy } from './devProxy'

afterEach(() => {
  vi.unstubAllEnvs()
})

describe('buildApiUrl', () => {
  it('uses the same-origin proxy prefix when API proxy is enabled', () => {
    expect(buildApiUrl('http://api.example.com/v1', 'images/edits', null, true)).toBe(
      '/api-proxy/images/edits',
    )
  })

  it('leaves API versioning to the proxy target when proxying', () => {
    expect(buildApiUrl('http://api.example.com', 'images/generations', null, true)).toBe(
      '/api-proxy/images/generations',
    )
  })

  it('uses a configured proxy prefix when one is available', () => {
    expect(
      buildApiUrl(
        'http://api.example.com/v1',
        'responses',
        {
          enabled: true,
          prefix: '/openai-proxy',
          target: 'http://api.example.com/v1',
          changeOrigin: true,
          secure: false,
        },
        true,
      ),
    ).toBe('/openai-proxy/responses')
  })

  it('uses the configured API URL directly when API proxy is disabled', () => {
    expect(buildApiUrl('http://api.example.com/v1', 'responses', null, false)).toBe(
      'http://api.example.com/v1/responses',
    )
  })

  it('disables the API proxy in fixed Codeingforce mode', () => {
    vi.stubEnv('VITE_SHOW_DEFAULT_CONFIG_ONLY', 'true')
    vi.stubEnv('VITE_API_PROXY_AVAILABLE', 'true')
    vi.stubEnv('VITE_API_PROXY_LOCKED', 'true')

    expect(isApiProxyAvailable()).toBe(false)
    expect(shouldUseApiProxy(true)).toBe(false)
  })
})
