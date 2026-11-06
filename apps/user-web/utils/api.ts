import { FetchError } from 'ofetch'

export class ApiError extends Error {
  status: number
  code: string
  requestId?: string

  constructor(status: number, code: string, message: string, requestId?: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.requestId = requestId
  }
}

export interface ApiOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  query?: Record<string, string | number | boolean | undefined>
  body?: unknown
  /** Send the access token. Defaults to true. */
  auth?: boolean
  /** Allow one silent refresh-and-retry on 401. Defaults to true. */
  retry?: boolean
}

let refreshPromise: Promise<string | null> | null = null

function readCookie(name: string): string | undefined {
  if (typeof document === 'undefined') return useCookie(name).value ?? undefined
  const match = document.cookie.match(new RegExp('(?:^|; )' + name + '=([^;]*)'))
  return match ? decodeURIComponent(match[1]) : undefined
}

function csrfHeaders(method: string): Record<string, string> {
  if (method === 'GET' || method === 'HEAD' || method === 'OPTIONS') return {}
  const token = readCookie('XSRF-TOKEN')
  if (!token) return {}
  return { 'X-CSRF-Token': token }
}

async function refreshAccess(): Promise<string | null> {
  if (!refreshPromise) {
    refreshPromise = (async () => {
      try {
        const config = useRuntimeConfig()
        const response = await $fetch<{ data: { access_token: string } }>('/api/v1/auth/refresh', {
          baseURL: config.public.apiBase,
          method: 'POST',
          credentials: 'include',
          headers: { ...csrfHeaders('POST') }
        })
        const token = response.data.access_token
        useCookie('tb_access', { maxAge: 900, sameSite: 'lax', path: '/' }).value = token
        return token
      } catch {
        useCookie('tb_access', { path: '/' }).value = null
        return null
      } finally {
        refreshPromise = null
      }
    })()
  }
  return refreshPromise
}

export function toApiError(error: unknown): ApiError {
  if (error instanceof ApiError) return error
  if (error instanceof FetchError) {
    const data = error.response?._data as { error?: { code?: string; message?: string; request_id?: string } } | undefined
    const status = error.response?.status ?? 0
    return new ApiError(
      status,
      data?.error?.code ?? 'REQUEST_FAILED',
      data?.error?.message ?? error.message,
      data?.error?.request_id
    )
  }
  return new ApiError(0, 'REQUEST_FAILED', error instanceof Error ? error.message : '요청에 실패했습니다')
}

export interface RawEnvelope<T> {
  data: T
  request_id: string
  next_cursor?: string
}

async function requestRaw<T>(path: string, options: ApiOptions, token?: string | null): Promise<RawEnvelope<T>> {
  const config = useRuntimeConfig()
  const method = options.method ?? 'GET'
  const headers: Record<string, string> = { ...csrfHeaders(method) }
  if (options.auth !== false && token) headers.Authorization = `Bearer ${token}`
  return await $fetch<RawEnvelope<T>>(path, {
    baseURL: config.public.apiBase,
    method,
    query: options.query as Record<string, string | number | boolean>,
    body: options.body as Record<string, unknown> | undefined,
    credentials: 'include',
    headers
  })
}

/** Full envelope (data + request_id + optional next_cursor). */
export async function apiRaw<T>(path: string, options: ApiOptions = {}): Promise<RawEnvelope<T>> {
  const withAuth = options.auth !== false
  try {
    const token = withAuth ? useCookie<string | null>('tb_access').value : undefined
    return await requestRaw<T>(path, options, token ?? undefined)
  } catch (error) {
    const status = error instanceof FetchError ? error.response?.status : undefined
    if (status === 401 && withAuth && options.retry !== false) {
      const fresh = await refreshAccess()
      if (fresh) {
        try {
          return await requestRaw<T>(path, options, fresh)
        } catch (retryError) {
          throw toApiError(retryError)
        }
      }
      const { clearSession } = useAuth()
      clearSession()
    }
    throw toApiError(error)
  }
}

export async function api<T>(path: string, options: ApiOptions = {}): Promise<T> {
  const config = useRuntimeConfig()
  const method = options.method ?? 'GET'
  const withAuth = options.auth !== false

  const attempt = async (token?: string | null): Promise<T> => {
    const headers: Record<string, string> = { ...csrfHeaders(method) }
    if (withAuth && token) headers.Authorization = `Bearer ${token}`
    const response = await $fetch<{ data: T }>(path, {
      baseURL: config.public.apiBase,
      method,
      query: options.query as Record<string, string | number | boolean>,
      body: options.body as Record<string, unknown> | undefined,
      credentials: 'include',
      headers
    })
    return response.data
  }

  try {
    const token = withAuth ? useCookie<string | null>('tb_access').value : undefined
    return await attempt(token ?? undefined)
  } catch (error) {
    const status = error instanceof FetchError ? error.response?.status : undefined
    if (status === 401 && withAuth && options.retry !== false) {
      const fresh = await refreshAccess()
      if (fresh) {
        try {
          return await attempt(fresh)
        } catch (retryError) {
          throw toApiError(retryError)
        }
      }
      const { clearSession } = useAuth()
      clearSession()
    }
    throw toApiError(error)
  }
}

const codeMessages: Record<string, string> = {
  DUPLICATE_EMAIL_OR_HANDLE: '이미 사용 중인 이메일 또는 사용자 이름입니다',
  INVALID_CREDENTIALS: '이메일 또는 비밀번호가 올바르지 않습니다',
  ACCOUNT_NOT_VERIFIED: '이메일 인증이 필요합니다',
  ACCOUNT_SUSPENDED: '정지된 계정입니다. 고객센터에 문의하세요',
  MFA_REQUIRED: 'MFA 인증 코드가 필요합니다',
  RATE_LIMIT_EXCEEDED: '요청이 너무 많습니다. 잠시 후 다시 시도하세요',
  POST_NOT_FOUND: '글을 찾을 수 없습니다',
  BLOG_NOT_FOUND: '블로그를 찾을 수 없습니다',
  COMMENTS_DISABLED: '댓글이 비활성화된 글입니다',
  COMMENTS_RESTRICTED: '이웃만 댓글을 작성할 수 있습니다',
  PAYMENTS_NOT_CONFIGURED: '결제 기능이 준비되지 않았습니다',
  OAUTH_NOT_CONFIGURED: '해당 소셜 로그인이 준비되지 않았습니다',
  OAUTH_EMAIL_REQUIRED: '제공자의 인증된 이메일이 필요합니다',
  CSRF_TOKEN_MISSING: '보안 토큰이 만료되었습니다. 새로고침 후 다시 시도하세요',
  CSRF_TOKEN_INVALID: '보안 토큰이 유효하지 않습니다. 새로고침 후 다시 시도하세요',
  SERVICE_UNAVAILABLE: '서버에 일시적인 문제가 있습니다. 잠시 후 다시 시도하세요'
}

export function friendlyMessage(error: unknown): string {
  const apiError = toApiError(error)
  return codeMessages[apiError.code] ?? apiError.message
}
