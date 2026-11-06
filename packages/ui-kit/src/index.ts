export const colors = {
  blue: {
    50: '#eef4ff',
    100: '#dbe6fd',
    200: '#bfd3fe',
    300: '#93b4fd',
    400: '#608dfa',
    500: '#3b6ef6',
    600: '#1d4ed8',
    700: '#1e40af',
    800: '#1e3a8a',
    900: '#172e6e'
  },
  ink: {
    400: '#8b98b0',
    500: '#64748f',
    600: '#40506a',
    700: '#24344f',
    900: '#101f38'
  },
  neutral: {
    50: '#f6f8fb',
    100: '#eef2f7',
    200: '#e4e9f0',
    300: '#c9d3e0',
    400: '#9aa7bc',
    500: '#64748f',
    600: '#40506a',
    700: '#24344f',
    800: '#16233a',
    900: '#101f38'
  }
} as const

export const theme = {
  primary: colors.blue[600],
  primaryStrong: colors.blue[700],
  accent: colors.blue[600],
  error: '#b42318',
  warning: '#8a5a00',
  info: colors.blue[600]
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
