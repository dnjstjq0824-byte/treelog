import { describe, expect, it } from 'vitest'
import {
  formatCount,
  formatDate,
  formatKRW,
  initials,
  isValidBlogHandle,
  isValidEmail,
  isValidUsername,
  passwordIssues,
  readingMinutes,
  timeAgo
} from './index'

describe('formatCount', () => {
  it('formats Korean units', () => {
    expect(formatCount(0)).toBe('0')
    expect(formatCount(999)).toBe('999')
    expect(formatCount(1500)).toBe('1.5천')
    expect(formatCount(123456)).toBe('12.34만')
  })
})

describe('formatKRW', () => {
  it('formats won', () => {
    expect(formatKRW(15000)).toBe('15,000원')
    expect(formatKRW(NaN)).toBe('0원')
  })
})

describe('formatDate', () => {
  it('handles invalid input', () => {
    expect(formatDate(null)).toBe('')
    expect(formatDate('not-a-date')).toBe('')
    expect(formatDate('2026-05-01T12:00:00Z')).toContain('2026')
  })
})

describe('timeAgo', () => {
  it('walks thresholds', () => {
    const now = new Date('2026-05-01T12:00:00Z').getTime()
    expect(timeAgo(new Date(now - 30 * 1000).toISOString(), now)).toBe('방금 전')
    expect(timeAgo(new Date(now - 5 * 60 * 1000).toISOString(), now)).toBe('5분 전')
    expect(timeAgo(new Date(now - 3 * 3600 * 1000).toISOString(), now)).toBe('3시간 전')
    expect(timeAgo(new Date(now - 2 * 86400 * 1000).toISOString(), now)).toBe('2일 전')
    expect(timeAgo('bad', now)).toBe('')
  })
})

describe('validators', () => {
  it('checks email', () => {
    expect(isValidEmail('a@b.co')).toBe(true)
    expect(isValidEmail('nope')).toBe(false)
  })

  it('lists password issues', () => {
    expect(passwordIssues('P@ssword123!')).toEqual([])
    expect(passwordIssues('short')).toContain('10자 이상 입력하세요')
  })

  it('checks username and handle', () => {
    expect(isValidUsername('kim-tree_9')).toBe(true)
    expect(isValidUsername('ab')).toBe(false)
    expect(isValidUsername('Bad Name')).toBe(false)
    expect(isValidBlogHandle('my-blog-1')).toBe(true)
    expect(isValidBlogHandle('Bad_Handle')).toBe(false)
  })
})

describe('misc', () => {
  it('computes reading time and initials', () => {
    expect(readingMinutes(1200)).toBe('2분 읽기')
    expect(initials('김 트리')).toBe('김트')
    expect(initials('')).toBe('?')
  })
})
