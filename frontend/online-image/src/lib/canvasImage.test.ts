import { afterEach, describe, expect, it, vi } from 'vitest'
import { dataUrlToBlob } from './canvasImage'

describe('canvas image helpers', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('converts data URLs to blobs without fetch', async () => {
    const fetchMock = vi.fn(() => {
      throw new Error('fetch should not be called')
    })
    vi.stubGlobal('fetch', fetchMock)

    const blob = await dataUrlToBlob('data:image/jpeg;base64,AQID')

    expect(fetchMock).not.toHaveBeenCalled()
    expect(blob.type).toBe('image/jpeg')
    expect(Array.from(new Uint8Array(await blob.arrayBuffer()))).toEqual([1, 2, 3])
  })
})
