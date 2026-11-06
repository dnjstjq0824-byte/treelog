/** API client for the TreeBlog backend (envelope, CSRF, refresh). */

export class ApiError extends Error {
  status: number
  code: string
  requestId?: string
  data?: unknown

  constructor(status: number, code: string, message: string, requestId?: string, data?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.requestId = requestId
    this.data = data
  }
}

export interface ApiOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  query?: Record<string, string | number | boolean | undefined>
  body?: unknown
  auth?: boolean
  retry?: boolean
}

export interface RawEnvelope<T> {
  data: T
  request_id: string
  next_cursor?: string
}

export function apiBase(): string {
  return import.meta.env.VITE_API_BASE || 'http://localhost:8080'
}

function readCookie(name: string): string | undefined {
  const match = document.cookie.match(new RegExp('(?:^|; )' + name + '=([^;]*)'))
  return match ? decodeURIComponent(match[1]) : undefined
}

export function accessToken(): string | null {
  return localStorage.getItem('tb_access')
}

export function saveAccess(token: string): void {
  localStorage.setItem('tb_access', token)
}

export function clearAccess(): void {
  localStorage.removeItem('tb_access')
}

function csrfHeaders(method: string): Record<string, string> {
  if (method === 'GET' || method === 'HEAD' || method === 'OPTIONS') return {}
  const token = readCookie('XSRF-TOKEN')
  if (!token) return {}
  return { 'X-CSRF-Token': token }
}

let refreshPromise: Promise<string | null> | null = null

async function refreshAccess(): Promise<string | null> {
  if (!refreshPromise) {
    refreshPromise = (async () => {
      try {
        const response = await fetch(`${apiBase()}/api/v1/auth/refresh`, {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json', ...csrfHeaders('POST') }
        })
        if (!response.ok) throw new Error('refresh failed')
        const envelope = (await response.json()) as RawEnvelope<{ access_token: string }>
        saveAccess(envelope.data.access_token)
        return envelope.data.access_token
      } catch {
        clearAccess()
        return null
      } finally {
        refreshPromise = null
      }
    })()
  }
  return refreshPromise
}

function buildURL(path: string, query?: ApiOptions['query']): string {
  const url = new URL(path, apiBase())
  if (query) {
    for (const [key, value] of Object.entries(query)) {
      if (value !== undefined) url.searchParams.set(key, String(value))
    }
  }
  return url.toString()
}

async function requestRaw<T>(path: string, options: ApiOptions, token?: string | null): Promise<RawEnvelope<T>> {
  const method = options.method ?? 'GET'
  const headers: Record<string, string> = { ...csrfHeaders(method) }
  if (options.auth !== false && token) headers.Authorization = `Bearer ${token}`
  const init: RequestInit = { method, credentials: 'include', headers }
  if (options.body !== undefined) {
    headers['Content-Type'] = 'application/json'
    init.body = JSON.stringify(options.body)
  }
  const response = await fetch(buildURL(path, options.query), init)
  const payload = (await response.json().catch(() => null)) as {
    data?: T
    error?: { code?: string; message?: string; request_id?: string; data?: unknown }
    request_id?: string
    next_cursor?: string
  } | null
  if (!response.ok) {
    throw new ApiError(
      response.status,
      payload?.error?.code ?? 'REQUEST_FAILED',
      payload?.error?.message ?? `요청 실패 (${response.status})`,
      payload?.error?.request_id,
      payload?.error?.data
    )
  }
  return {
    data: payload?.data as T,
    request_id: payload?.request_id ?? '',
    next_cursor: payload?.next_cursor
  }
}

export async function apiRaw<T>(path: string, options: ApiOptions = {}): Promise<RawEnvelope<T>> {
  const withAuth = options.auth !== false
  try {
    return await requestRaw<T>(path, options, withAuth ? accessToken() : undefined)
  } catch (error) {
    if (error instanceof ApiError && error.status === 401 && withAuth && options.retry !== false) {
      const fresh = await refreshAccess()
      if (fresh) return await requestRaw<T>(path, options, fresh)
    }
    throw error
  }
}

export async function api<T>(path: string, options: ApiOptions = {}): Promise<T> {
  return (await apiRaw<T>(path, options)).data
}

const codeMessages: Record<string, string> = {
  DUPLICATE_EMAIL_OR_HANDLE: '이미 사용 중인 이메일 또는 사용자 이름입니다',
  INVALID_CREDENTIALS: '이메일 또는 비밀번호가 올바르지 않습니다',
  ACCOUNT_NOT_VERIFIED: '이메일 인증이 필요합니다',
  ACCOUNT_SUSPENDED: '정지된 계정입니다. 고객센터에 문의하세요',
  RATE_LIMIT_EXCEEDED: '요청이 너무 많습니다. 잠시 후 다시 시도하세요',
  POST_NOT_FOUND: '글을 찾을 수 없습니다',
  BLOG_NOT_FOUND: '블로그를 찾을 수 없습니다',
  FORBIDDEN: '이 블로그의 멤버가 아닙니다',
  INSUFFICIENT_PERMISSIONS: '권한이 부족합니다',
  INVALID_STATUS: '알 수 없는 글 상태입니다',
  AUTOSAVE_CONFLICT: '다른 기기에서 수정된 내용이 있습니다',
  SERVICE_UNAVAILABLE: '서버에 일시적인 문제가 있습니다. 잠시 후 다시 시도하세요'
}

export function friendlyMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return codeMessages[error.code] ?? error.message
  }
  return error instanceof Error ? error.message : '요청에 실패했습니다'
}
