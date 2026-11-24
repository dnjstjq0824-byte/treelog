import { describe, expect, it } from 'vitest'
import { buildSrcset, publicMediaURL, type MediaVariantInfo } from './media'

const variant = (overrides: Partial<MediaVariantInfo>): MediaVariantInfo => ({
  variant_name: 'W640',
  format: 'JPEG',
  width: 640,
  height: 480,
  size_bytes: 1000,
  public_cdn_url: '/x',
  ...overrides
})

describe('publicMediaURL', () => {
  it('builds variant and format URLs', () => {
    expect(publicMediaURL('a1')).toContain('/api/v1/public/media/a1')
    expect(publicMediaURL('a1', 'THUMB_400')).toContain('variant=THUMB_400')
    expect(publicMediaURL('a1', 'W640', 'webp')).toContain('format=webp')
  })
})

describe('buildSrcset', () => {
  it('prefers WebP twins and sorts by width', () => {
    const srcset = buildSrcset('a1', [
      variant({ variant_name: 'W640', format: 'JPEG', width: 640 }),
      variant({ variant_name: 'W320', format: 'JPEG', width: 320 }),
      variant({ variant_name: 'THUMB_400', format: 'JPEG', width: 400 }),
      variant({ variant_name: 'W640', format: 'WEBP', width: 640 })
    ])
    const entries = srcset.split(', ')
    expect(entries).toHaveLength(2)
    expect(entries[0]).toContain('320w')
    expect(entries[1]).toContain('640w')
    expect(entries[1]).toContain('format=webp')
  })

  it('returns empty without width variants', () => {
    expect(buildSrcset('a1', [])).toBe('')
    expect(buildSrcset('a1', [variant({ variant_name: 'THUMB_400', width: 400 })])).toBe('')
  })
})
