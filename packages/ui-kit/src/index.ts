export const colors = {
  pine: {
    50: '#eff7f1',
    100: '#dcefe2',
    200: '#bbdfc6',
    300: '#8fc7a2',
    400: '#5daa78',
    500: '#3d8b5c',
    600: '#2a7049',
    700: '#235c3d',
    800: '#1d4a32',
    900: '#0e2e21'
  },
  clay: {
    100: '#fbe9e2',
    500: '#d0653f',
    600: '#c2543b',
    700: '#a0432d'
  },
  neutral: {
    50: '#fafaf8',
    100: '#f2f1ec',
    200: '#e6e4dc',
    300: '#d5d2c7',
    400: '#a8a49a',
    500: '#7c786e',
    600: '#5f5b52',
    700: '#4a4740',
    800: '#34322d',
    900: '#1e1d1a'
  }
} as const

export const theme = {
  primary: colors.pine[600],
  primaryStrong: colors.pine[700],
  accent: colors.clay[600],
  error: '#b3261e',
  warning: '#b45309',
  info: '#1d5bbf'
} as const

export function formatCount(value: number): string {
  if (!Number.isFinite(value) || value < 0) return '0'
  if (value >= 10000) {
    const rounded = Math.floor(value / 100) / 100
    return `${trimZeros(rounded)}만`
  }
  if (value >= 1000) {
    const rounded = Math.floor(value / 100) / 10
    return `${trimZeros(rounded)}천`
  }
  return String(Math.floor(value))
}

function trimZeros(value: number): string {
  return String(Math.round(value * 100) / 100)
}

export function formatKRW(minor: number): string {
  if (!Number.isFinite(minor)) return '0원'
  return `${Math.round(minor).toLocaleString('ko-KR')}원`
}

export function formatDate(iso: string | null | undefined): string {
  if (!iso) return ''
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleDateString('ko-KR', { year: 'numeric', month: 'long', day: 'numeric' })
}

export function formatDateTime(iso: string | null | undefined): string {
  if (!iso) return ''
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleString('ko-KR', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

export function timeAgo(iso: string | null | undefined, now: number = Date.now()): string {
  if (!iso) return ''
  const then = new Date(iso).getTime()
  if (Number.isNaN(then)) return ''
  const seconds = Math.max(0, Math.floor((now - then) / 1000))
  if (seconds < 60) return '방금 전'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}분 전`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}시간 전`
  const days = Math.floor(hours / 24)
  if (days < 7) return `${days}일 전`
  if (days < 30) return `${Math.floor(days / 7)}주 전`
  if (days < 365) return `${Math.floor(days / 30)}개월 전`
  return `${Math.floor(days / 365)}년 전`
}

export function readingMinutes(wordCount: number): string {
  if (!Number.isFinite(wordCount) || wordCount <= 0) return '1분 읽기'
  return `${Math.max(1, Math.round(wordCount / 500))}분 읽기`
}

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export function isValidEmail(email: string): boolean {
  return emailPattern.test(email.trim())
}

export function passwordIssues(password: string): string[] {
  const issues: string[] = []
  if (password.length < 10) issues.push('10자 이상 입력하세요')
  if (!/[a-zA-Z]/.test(password)) issues.push('영문을 포함하세요')
  if (!/[0-9]/.test(password)) issues.push('숫자를 포함하세요')
  if (!/[^a-zA-Z0-9]/.test(password)) issues.push('특수문자를 포함하세요')
  return issues
}

export function isValidUsername(username: string): boolean {
  return /^[a-z0-9][a-z0-9_-]{2,31}$/.test(username)
}

export function isValidBlogHandle(handle: string): boolean {
  return /^[a-z0-9][a-z0-9-]{1,28}[a-z0-9]$/.test(handle)
}

export function initials(name: string): string {
  const trimmed = name.trim()
  if (!trimmed) return '?'
  const parts = trimmed.split(/\s+/)
  if (parts.length === 1) return trimmed.slice(0, 2).toUpperCase()
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
}
