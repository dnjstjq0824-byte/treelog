<template>
  <div>
    <h1 class="tb-section-title">프로필 설정</h1>
    <form @submit.prevent="save" style="max-width: 520px">
      <div class="tb-field">
        <label class="tb-label" for="me-display">표시 이름</label>
        <input id="me-display" v-model="displayName" class="tb-input" maxlength="50" required />
      </div>
      <div class="tb-field">
        <label class="tb-label" for="me-bio">소개</label>
        <textarea
          id="me-bio"
          v-model="bio"
          class="tb-textarea"
          rows="3"
          maxlength="500"
          placeholder="자신을 소개해 보세요"
        />
      </div>
      <p v-if="message" class="tb-error-text" role="status">{{ message }}</p>
      <div class="tb-form-actions">
        <button class="tb-btn tb-btn-primary" :disabled="busy">
          {{ busy ? '저장 중…' : '저장' }}
        </button>
      </div>
    </form>

    <h2 class="tb-section-title" style="margin-top: 32px">2단계 인증 (TOTP)</h2>
    <div style="max-width: 520px">
      <div class="tb-list-item">
        <div>
          <p class="tb-list-item-title">인증 앱</p>
          <p class="tb-list-item-desc">{{ mfaEnabled ? '사용 중' : '사용 안 함' }}</p>
        </div>
        <button
          v-if="!mfaEnabled"
          class="tb-btn tb-btn-secondary tb-btn-sm"
          @click="toggleMfaPanel"
        >
          {{ showMfa ? '닫기' : '설정' }}
        </button>
        <button v-else class="tb-btn tb-btn-secondary tb-btn-sm" @click="toggleDisablePanel">
          {{ showDisable ? '닫기' : '해제' }}
        </button>
      </div>
      <form v-if="showMfa" @submit.prevent="confirmMfa" style="margin-top: 12px">
        <p class="tb-hint">인증 앱에서 아래 주소를 등록한 뒤 6자리 코드를 입력하세요.</p>
        <code class="tb-code">{{ mfaUri || '생성 중…' }}</code>
        <p v-if="backupCodes.length > 0" class="tb-hint" style="margin-top: 8px">
          복구 코드 (안전한 곳에 보관하세요): {{ backupCodes.join(', ') }}
        </p>
        <div class="tb-field" style="margin-top: 12px">
          <label class="tb-label" for="mfa-code">6자리 코드</label>
          <input
            id="mfa-code"
            v-model="mfaCode"
            class="tb-input"
            inputmode="numeric"
            maxlength="6"
            autocomplete="one-time-code"
          />
        </div>
        <div class="tb-form-actions">
          <button class="tb-btn tb-btn-primary tb-btn-sm" :disabled="mfaBusy">
            {{ mfaBusy ? '확인 중…' : '확인' }}
          </button>
        </div>
      </form>
      <form v-if="showDisable" @submit.prevent="disableMfa" style="margin-top: 12px">
        <div class="tb-field">
          <label class="tb-label" for="mfa-disable-code">현재 6자리 코드</label>
          <input
            id="mfa-disable-code"
            v-model="disableCode"
            class="tb-input"
            inputmode="numeric"
            maxlength="6"
            autocomplete="one-time-code"
          />
        </div>
        <div class="tb-form-actions">
          <button class="tb-btn tb-btn-danger tb-btn-sm" :disabled="mfaBusy">
            {{ mfaBusy ? '해제 중…' : '2단계 인증 해제' }}
          </button>
        </div>
      </form>
    </div>

    <h2 class="tb-section-title" style="margin-top: 32px">로그인 세션</h2>
    <div style="max-width: 520px">
      <div v-if="sessionsPending">
        <div v-for="index in 2" :key="index" class="tb-skeleton" style="height: 52px" />
      </div>
      <ul v-else class="tb-list">
        <li v-for="session in sessions" :key="session.id" class="tb-list-item">
          <div>
            <p class="tb-list-item-title">
              {{ session.is_current ? '이 기기' : '다른 기기' }}
              <span v-if="session.is_current" class="tb-badge">현재</span>
            </p>
            <p class="tb-list-item-desc">최근 활동 {{ timeAgo(session.last_active_at) }}</p>
          </div>
          <button
            v-if="!session.is_current"
            class="tb-btn tb-btn-ghost tb-btn-sm"
            @click="revokeSession(session.id)"
          >
            해제
          </button>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { timeAgo } from '@treeblog/ui-kit'
import type { MfaSetup, SessionInfo } from '~/types/api'
import { api, friendlyMessage } from '~/utils/api'

useSeoMeta({ title: '프로필 설정 | TreeBlog', robots: 'noindex' })
definePageMeta({ middleware: 'auth' })

const { user, refresh } = useAuth()
const toast = useToast()
const displayName = ref('')
const bio = ref('')
const busy = ref(false)
const message = ref('')
const mfaEnabled = ref(false)
const showMfa = ref(false)
const showDisable = ref(false)
const mfaUri = ref('')
const mfaSecret = ref('')
const backupCodes = ref<string[]>([])
const mfaCode = ref('')
const disableCode = ref('')
const mfaBusy = ref(false)
const sessions = ref<SessionInfo[]>([])
const sessionsPending = ref(true)

watchEffect(() => {
  if (user.value) {
    displayName.value = user.value.display_name
    bio.value = user.value.bio || ''
    mfaEnabled.value = user.value.mfa_enabled
  }
})

async function save(): Promise<void> {
  busy.value = true
  message.value = ''
  try {
    await api('/api/v1/me/profile', {
      method: 'PATCH',
      body: { display_name: displayName.value.trim(), bio: bio.value.trim() }
    })
    await refresh()
    toast.push('프로필을 저장했습니다', 'success')
  } catch (err) {
    message.value = friendlyMessage(err)
  } finally {
    busy.value = false
  }
}

async function toggleMfaPanel(): Promise<void> {
  showMfa.value = !showMfa.value
  if (!showMfa.value) return
  mfaUri.value = ''
  mfaCode.value = ''
  backupCodes.value = []
  try {
    const data = await api<MfaSetup>('/api/v1/auth/mfa/setup', { method: 'POST' })
    mfaSecret.value = data.secret
    backupCodes.value = data.backup_codes
    const label = encodeURIComponent(`TreeBlog:${user.value?.email || 'user'}`)
    mfaUri.value = `otpauth://totp/${label}?secret=${data.secret}&issuer=TreeBlog`
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
    showMfa.value = false
  }
}

async function confirmMfa(): Promise<void> {
  mfaBusy.value = true
  try {
    await api('/api/v1/auth/mfa/verify', {
      method: 'POST',
      body: { code: mfaCode.value.trim() }
    })
    mfaEnabled.value = true
    showMfa.value = false
    await refresh()
    toast.push('2단계 인증을 설정했습니다', 'success')
  } catch {
    toast.push('코드가 올바르지 않습니다', 'error')
  } finally {
    mfaBusy.value = false
  }
}

function toggleDisablePanel(): void {
  showDisable.value = !showDisable.value
  disableCode.value = ''
}

async function disableMfa(): Promise<void> {
  mfaBusy.value = true
  try {
    await api('/api/v1/auth/mfa/', {
      method: 'DELETE',
      body: { code: disableCode.value.trim() }
    })
    mfaEnabled.value = false
    showDisable.value = false
    await refresh()
    toast.push('2단계 인증을 해제했습니다', 'success')
  } catch {
    toast.push('코드가 올바르지 않습니다', 'error')
  } finally {
    mfaBusy.value = false
  }
}

async function revokeSession(sessionId: string): Promise<void> {
  try {
    await api(`/api/v1/me/sessions/${sessionId}`, { method: 'DELETE' })
    sessions.value = sessions.value.filter((session) => session.id !== sessionId)
    toast.push('세션을 해제했습니다', 'success')
  } catch {
    toast.push('세션 해제에 실패했습니다', 'error')
  }
}

onMounted(async () => {
  try {
    sessions.value = await api<SessionInfo[]>('/api/v1/me/sessions')
  } catch {
    sessions.value = []
  } finally {
    sessionsPending.value = false
  }
})
</script>
