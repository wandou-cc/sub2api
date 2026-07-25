import { afterEach, describe, expect, it, vi } from 'vitest'
import { downloadImageIds } from './downloadImages'

vi.mock('../store', () => ({
  ensureImageCached: vi.fn(),
}))

describe('download images', () => {
  afterEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('downloads data URL images without fetch', async () => {
    const fetchMock = vi.fn(() => {
      throw new Error('fetch should not be called')
    })
    const anchor = { href: '', download: '', click: vi.fn() }
    const createObjectURL = vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:download')
    vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => {})
    vi.stubGlobal('fetch', fetchMock)
    vi.stubGlobal('document', {
      createElement: vi.fn(() => anchor),
      body: {
        appendChild: vi.fn(),
        removeChild: vi.fn(),
      },
    })
    vi.stubGlobal('window', { setTimeout: vi.fn() })

    const result = await downloadImageIds(['data:image/png;base64,AQID'], 'image')

    expect(result).toEqual({ successCount: 1, failCount: 0 })
    expect(fetchMock).not.toHaveBeenCalled()
    expect(anchor.download).toBe('image.png')
    expect(anchor.click).toHaveBeenCalledOnce()
    const blob = createObjectURL.mock.calls[0][0] as Blob
    expect(blob.type).toBe('image/png')
    expect(Array.from(new Uint8Array(await blob.arrayBuffer()))).toEqual([1, 2, 3])
  })
})
