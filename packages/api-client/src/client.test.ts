import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiClientError, createTreeblogClient, newIdempotencyKey, throwApiError } from './client'

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' }
  })
}

describe('createTreeblogClient', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('injects auth, request id, csrf, and idempotency headers', async () => {
    const seen: Request[] = []
    const stubFetch = vi.fn(async (input: RequestInfo | URL): Promise<Response> => {
      seen.push(input as Request)
      return jsonResponse(200, { data: { ok: true }, request_id: 'r1' })
    })
    const client = createTreeblogClient({
      baseUrl: 'https://api.example.com/',
      getAccessToken: () => 'tok',
      getCsrfToken: () => 'csrf',
      generateRequestId: () => 'req-1',
      fetch: stubFetch as typeof fetch
    })
    const { data, error } = await client.POST('/api/v1/comments/{commentId}/approve', {
      params: { path: { commentId: 'c1' } }
    })
    expect(error).toBeUndefined()
    expect(data).toBeTruthy()
    expect(seen).toHaveLength(1)
    const request = seen[0] as Request
    expect(request.url).toBe('https://api.example.com/api/v1/comments/c1/approve')
    expect(request.headers.get('Authorization')).toBe('Bearer tok')
    expect(request.headers.get('X-Request-Id')).toBe('req-1')
    expect(request.headers.get('X-CSRF-Token')).toBe('csrf')
    expect(request.headers.get('Idempotency-Key')).toBeTruthy()
  })

  it('refreshes once on 401 and retries with the new token', async () => {
    const seen: Request[] = []
    const stubFetch = vi.fn(async (input: RequestInfo | URL): Promise<Response> => {
      const request = input as Request
      seen.push(request)
      if (request.headers.get('Authorization') === 'Bearer fresh') {
        return jsonResponse(200, { data: [], request_id: 'r2' })
      }
      return jsonResponse(401, { error: { code: 'AUTH_TOKEN_EXPIRED', message: 'expired' } })
    })
    const client = createTreeblogClient({
      baseUrl: 'https://api.example.com',
      getAccessToken: () => 'stale',
      onUnauthorized: async () => 'fresh',
      fetch: stubFetch as typeof fetch
    })
    const { data, error } = await client.GET('/api/v1/search/popular-queries', {
      params: { query: { limit: 5 } }
    })
    expect(error).toBeUndefined()
    expect(data).toBeTruthy()
    expect(seen).toHaveLength(2)
    expect(seen[1]?.headers.get('Authorization')).toBe('Bearer fresh')
  })

  it('does not retry twice when refresh fails', async () => {
    const stubFetch = vi.fn(async (): Promise<Response> =>
      jsonResponse(401, { error: { code: 'AUTH_TOKEN_EXPIRED', message: 'expired' } })
    )
    const client = createTreeblogClient({
      baseUrl: 'https://api.example.com',
      getAccessToken: () => 'stale',
      onUnauthorized: async () => null,
      fetch: stubFetch as typeof fetch
    })
    const { error } = await client.GET('/api/v1/search/popular-queries', {})
    expect(error).toBeTruthy()
    expect(stubFetch).toHaveBeenCalledTimes(1)
  })
})

describe('throwApiError', () => {
  it('parses the standard envelope', () => {
    try {
      throwApiError(403, { error: { code: 'POST_PRIVATE', message: '비공개', request_id: 'r9' } })
      expect.unreachable()
    } catch (error) {
      expect(error).toBeInstanceOf(ApiClientError)
      const apiError = error as ApiClientError
      expect(apiError.status).toBe(403)
      expect(apiError.code).toBe('POST_PRIVATE')
      expect(apiError.requestId).toBe('r9')
    }
  })

  it('falls back for unknown bodies', () => {
    try {
      throwApiError(500, null)
      expect.unreachable()
    } catch (error) {
      const apiError = error as ApiClientError
      expect(apiError.code).toBe('REQUEST_FAILED')
    }
  })
})

describe('newIdempotencyKey', () => {
  it('generates unique keys', () => {
    expect(newIdempotencyKey()).toBeTruthy()
    expect(newIdempotencyKey()).not.toBe(newIdempotencyKey())
  })
})
