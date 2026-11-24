import { describe, expect, it } from 'vitest'
import { MemoryDraftStore, shouldRestoreLocal, type LocalDraft } from './drafts'

const draft = (overrides: Partial<LocalDraft> = {}): LocalDraft => ({
  postId: 'p1',
  blogId: 'b1',
  title: '제목',
  subtitle: '',
  content: { type: 'doc', content: [] },
  clientTime: '2026-11-24T10:00:00.000Z',
  serverRevisionNo: 3,
  ...overrides
})

describe('MemoryDraftStore', () => {
  it('round-trips drafts and isolates snapshots from later mutations', async () => {
    const store = new MemoryDraftStore()
    expect(await store.load('p1')).toBeNull()
    const original = draft()
    await store.save(original)
    original.title = '바뀐 제목'
    const loaded = await store.load('p1')
    expect(loaded?.title).toBe('제목')
    await store.clear('p1')
    expect(await store.load('p1')).toBeNull()
  })
})

describe('shouldRestoreLocal', () => {
  it('restores only newer caches on the same revision base', () => {
    expect(shouldRestoreLocal(null, 3, null)).toBe(false)
    expect(shouldRestoreLocal(draft(), 3, null)).toBe(true)
    expect(shouldRestoreLocal(draft(), 3, '2026-11-24T09:00:00.000Z')).toBe(true)
    expect(shouldRestoreLocal(draft(), 3, '2026-11-24T10:00:00.000Z')).toBe(false)
    expect(shouldRestoreLocal(draft(), 3, '2026-11-24T11:00:00.000Z')).toBe(false)
    expect(shouldRestoreLocal(draft(), 4, null)).toBe(false)
  })
})
