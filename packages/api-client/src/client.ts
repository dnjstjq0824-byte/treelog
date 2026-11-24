import createClient, { type Middleware } from 'openapi-fetch'
import type { paths } from './schema'

export type { paths }

// StandardErrorResponse mirrors 04_API/API_ERROR_MODEL.md.
export interface StandardError {
  code: string
  message: string
  request_id?: string
  timestamp?: string
}

export class ApiClientError extends Error {
  readonly status: number
  readonly code: string
  readonly requestId?: string

  constructor(status: number, code: string, message: string, requestId?: string) {
    super(message)
    this.name = 'ApiClientError'
    this.status = status
    this.code = code
    this.requestId = requestId
  }
}

function randomId(): string {
  const cryptoRef =
    typeof globalThis.crypto !== 'undefined' ? (globalThis.crypto as Crypto & { randomUUID?: () => string }) : undefined
  if (cryptoRef?.randomUUID) return cryptoRef.randomUUID()
  return `id-${Date.now().toString(36)}-${Math.floor(Math.random() * 1e9).toString(36)}`
}

// newIdempotencyKey generates the Idempotency-Key value for mutating calls.
export function newIdempotencyKey(): string {
  return randomId()
}

export interface TreeblogClientOptions {
  baseUrl: string
  getAccessToken?: () => string | null | undefined
  /** Refresh hook: returns a new access token, or null when re-login is needed. */
  onUnauthorized?: () => Promise<string | null>
  getCsrfToken?: () => string | null | undefined
  generateRequestId?: () => string
  /** Test seam for the underlying fetch implementation. */
  fetch?: typeof fetch
}

export type TreeblogClient = ReturnType<typeof createTreeblogClient>

export function createTreeblogClient(options: TreeblogClientOptions) {
  const generateRequestId = options.generateRequestId ?? randomId
  const client = createClient<paths>({
    baseUrl: options.baseUrl.replace(/\/$/, ''),
    fetch: options.fetch
  })

  const auth: Middleware = {
    async onRequest({ request }) {
      const token = options.getAccessToken?.()
      if (token) request.headers.set('Authorization', `Bearer ${token}`)
      request.headers.set('X-Request-Id', generateRequestId())
      if (request.method !== 'GET' && request.method !== 'HEAD' && request.method !== 'OPTIONS') {
        const csrf = options.getCsrfToken?.()
        if (csrf) request.headers.set('X-CSRF-Token', csrf)
        if (!request.headers.has('Idempotency-Key')) {
          request.headers.set('Idempotency-Key', newIdempotencyKey())
        }
      }
      return request
    },
    async onResponse({ request, response }) {
      if (response.status !== 401 || !options.onUnauthorized) return response
      if (request.headers.get('X-Tb-Retried') === '1') return response
      const fresh = await options.onUnauthorized().catch(() => null)
      if (!fresh) return response
      const retry = request.clone()
      retry.headers.set('Authorization', `Bearer ${fresh}`)
      retry.headers.set('X-Tb-Retried', '1')
      const impl = options.fetch ?? fetch
      return impl(retry)
    }
  }
  client.use(auth)
  return client
}

// throwApiError converts a failed openapi-fetch response into ApiClientError.
// Use it after calls: `const { data, error, response } = await client.GET(...); if (error) throwApiError(response.status, error)`.
export function throwApiError(status: number, body: unknown): never {
  const fallback = '요청에 실패했습니다'
  if (body && typeof body === 'object' && 'error' in body) {
    const inner = (body as { error?: Partial<StandardError> }).error ?? {}
    throw new ApiClientError(
      status,
      typeof inner.code === 'string' ? inner.code : 'REQUEST_FAILED',
      typeof inner.message === 'string' ? inner.message : fallback,
      typeof inner.request_id === 'string' ? inner.request_id : undefined
    )
  }
  throw new ApiClientError(status, 'REQUEST_FAILED', fallback)
}
