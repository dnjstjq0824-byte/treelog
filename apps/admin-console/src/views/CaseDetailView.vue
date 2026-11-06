<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">케이스 상세</h1>
      <RouterLink to="/moderation" class="tb-btn tb-btn-ghost tb-btn-sm">목록으로</RouterLink>
    </div>
    <div v-if="pending">
      <div class="tb-skeleton" style="height: 300px" />
    </div>
    <div v-else-if="detail">
      <div class="tb-list-item" style="border: 1px solid var(--tb-border); border-radius: var(--tb-radius-lg); margin-bottom: 16px">
        <div>
          <p class="tb-list-item-title">{{ detail.case.target_type }} · {{ detail.case.reason_category }}</p>
          <p class="tb-list-item-desc">
            대상 {{ detail.case.target_id }} · 신고 {{ detail.case.report_count }}건 ·
            우선순위 {{ detail.case.priority_score }}
          </p>
        </div>
        <span class="tb-badge">{{ detail.case.status }}</span>
      </div>

      <h2 class="tb-section-title">신고 내역</h2>
      <ul class="tb-list">
        <li v-for="report in detail.reports" :key="report.id" class="tb-list-item">
          <div>
            <p class="tb-list-item-title">{{ report.reason_category }}</p>
            <p class="tb-list-item-desc">{{ report.description || '설명 없음' }}</p>
            <p class="tb-hint">{{ formatDateTime(report.created_at) }}</p>
          </div>
        </li>
      </ul>

      <h2 class="tb-section-title" style="margin-top: 20px">조치 내역</h2>
      <ul v-if="detail.actions.length > 0" class="tb-list">
        <li v-for="action in detail.actions" :key="action.id" class="tb-list-item">
          <div>
            <p class="tb-list-item-title">{{ actionLabel(action.action_type) }}</p>
            <p class="tb-list-item-desc">
              {{ action.policy_clause_code }} · {{ action.public_notice_reason }}
            </p>
            <p class="tb-hint">{{ formatDateTime(action.created_at) }}</p>
          </div>
        </li>
      </ul>
      <p v-else class="tb-hint">아직 조치 내역이 없습니다.</p>

      <h2 class="tb-section-title" style="margin-top: 20px">조치 적용</h2>
      <form style="max-width: 560px" @submit.prevent="applyAction">
        <div class="tb-field">
          <label class="tb-label" for="action-type">조치 종류</label>
          <select id="action-type" v-model="form.action_type" class="tb-select">
            <option v-for="option in ACTION_TYPES" :key="option.id" :value="option.id">
              {{ option.label }}
            </option>
          </select>
        </div>
        <div class="tb-field">
          <label class="tb-label" for="action-clause">정책 조항 코드 (필수, 40자 이내)</label>
          <input id="action-clause" v-model="form.policy_clause_code" class="tb-input" required maxlength="40" placeholder="예: 6.1" />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="action-note">내부 메모</label>
          <input id="action-note" v-model="form.internal_note" class="tb-input" maxlength="500" />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="action-public">공개 사유 (필수, 500자 이내)</label>
          <textarea id="action-public" v-model="form.public_notice_reason" class="tb-textarea" rows="2" required maxlength="500" />
        </div>
        <div class="tb-form-actions">
          <button class="tb-btn tb-btn-primary" :disabled="busy">
            {{ busy ? '적용 중…' : '조치 적용' }}
          </button>
        </div>
      </form>
    </div>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">케이스를 찾을 수 없습니다</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatDateTime } from '@treeblog/ui-kit'
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, friendlyMessage } from '../lib/api'
import { ACTION_TYPES, type ActionRow, type CaseDetail } from '../lib/types'
import { useToastStore } from '../stores/toast'

const route = useRoute()
const toast = useToastStore()
const caseId = computed(() => String(route.params.caseId || ''))

const detail = ref<CaseDetail | null>(null)
const pending = ref(true)
const busy = ref(false)
const form = ref({
  action_type: 'WARN_USER',
  policy_clause_code: '',
  internal_note: '',
  public_notice_reason: ''
})

function actionLabel(value: string): string {
  return ACTION_TYPES.find((option) => option.id === value)?.label || value
}

async function load(): Promise<void> {
  pending.value = true
  try {
    detail.value = await api<CaseDetail>(`/api/v1/admin/moderation/cases/${caseId.value}`)
  } catch (err) {
    detail.value = null
    toast.push(friendlyMessage(err), 'error')
  } finally {
    pending.value = false
  }
}

async function applyAction(): Promise<void> {
  busy.value = true
  try {
    const row = await api<ActionRow>(`/api/v1/admin/moderation/cases/${caseId.value}/actions`, {
      method: 'POST',
      body: {
        action_type: form.value.action_type,
        policy_clause_code: form.value.policy_clause_code.trim(),
        internal_note: form.value.internal_note.trim(),
        public_notice_reason: form.value.public_notice_reason.trim()
      }
    })
    if (detail.value) detail.value.actions = [...detail.value.actions, row]
    form.value = { action_type: 'WARN_USER', policy_clause_code: '', internal_note: '', public_notice_reason: '' }
    toast.push('조치를 적용했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    busy.value = false
  }
}

onMounted(load)
</script>
