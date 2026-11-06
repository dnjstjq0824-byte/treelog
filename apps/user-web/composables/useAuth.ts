import type { RegisterResponse, SessionUser, Tokens } from '~/types/api'
import { api } from '~/utils/api'

const oauthErrorMessages: Record<string, string> = {
  OAUTH_PROVIDER_DENIED: '소셜 로그인이 취소되었습니다',
  INVALID_OAUTH_STATE: '로그인 세션이 만료되었습니다. 다시 시도해 주세요',
  OAUTH_NOT_CONFIGURED: '해당 소셜 로그인이 준비되지 않았습니다',
  OAUTH_EMAIL_REQUIRED: '제공자의 인증된 이메일이 필요합니다',
  OAUTH_FAILED: '소셜 로그인에 실패했습니다'
}

export interface RegisterPayload {
  email: string
  username: string
  display_name: string
  password: string
}

export function useAuth() {
  const user = useState<SessionUser | null>('tb_user', () => null)
  const initialized = useState('tb_auth_init', () => false)
  const toast = useToast()

  async function fetchMe(): Promise<SessionUser | null> {
    try {
      const me = await api<SessionUser>('/api/v1/me/')
      user.value = me
      return me
    } catch {
      user.value = null
      return null
    }
  }

  async function initialize(): Promise<void> {
    if (initialized.value) return
    initialized.value = true
    const token = useCookie<string | null>('tb_access').value
    if (token) await fetchMe()
  }

  function saveAccess(token: string, expiresIn: number): void {
    useCookie('tb_access', {
      maxAge: Math.max(60, expiresIn - 30),
      sameSite: 'lax',
      path: '/'
    }).value = token
  }

  function clearSession(): void {
    user.value = null
    useCookie('tb_access', { path: '/' }).value = null
  }

  async function login(email: string, password: string): Promise<void> {
    const tokens = await api<Tokens>('/api/v1/auth/login', {
      method: 'POST',
      auth: false,
      body: { email, password }
    })
    saveAccess(tokens.access_token, tokens.expires_in)
    await fetchMe()
    toast.push('로그인했습니다', 'success')
    await navigateTo('/')
  }

  async function register(payload: RegisterPayload): Promise<void> {
    const result = await api<RegisterResponse>('/api/v1/auth/register', {
      method: 'POST',
      auth: false,
      body: { ...payload, terms_version: '1.0', privacy_version: '1.0' }
    })
    toast.push('인증 코드를 이메일로 보냈습니다', 'success')
    await navigateTo({ path: '/verify-email', query: { email: result.email } })
  }

  async function logout(): Promise<void> {
    try {
      await api('/api/v1/auth/logout', { method: 'POST' })
    } catch {
      // Session is gone server-side or already expired; clear locally anyway.
    }
    clearSession()
    await navigateTo('/')
  }

  async function startOAuth(provider: string): Promise<void> {
    const config = useRuntimeConfig()
    await navigateTo(`${config.public.apiBase}/api/v1/auth/oauth/${provider}/login`, {
      external: true
    })
  }

  /** Backend redirects here with `#access_token=…&expires_in=…` or `?error=…`. */
  async function finishOAuth(): Promise<void> {
    const route = useRoute()
    const queryError = route.query.error
    if (typeof queryError === 'string' && queryError) {
      throw new Error(oauthErrorMessages[queryError] || '소셜 로그인에 실패했습니다')
    }
    const fragment = new URLSearchParams(window.location.hash.replace(/^#/, ''))
    const accessToken = fragment.get('access_token')
    const expiresIn = Number(fragment.get('expires_in') || '900')
    if (!accessToken) {
      throw new Error('소셜 로그인 정보를 받지 못했습니다')
    }
    saveAccess(accessToken, Number.isFinite(expiresIn) ? expiresIn : 900)
    window.location.hash = ''
    await fetchMe()
    toast.push('로그인했습니다', 'success')
    await navigateTo('/')
  }

  return {
    user: readonly(user),
    isLoggedIn: computed(() => user.value !== null),
    initialized: readonly(initialized),
    initialize,
    fetchMe,
    refresh: fetchMe,
    clearSession,
    login,
    register,
    logout,
    startOAuth,
    finishOAuth
  }
}
