<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">도메인</h1>
      <button class="tb-btn tb-btn-primary tb-btn-sm" @click="showRegister = true">도메인 등록</button>
    </div>

    <div v-if="showRegister" class="tb-modal-backdrop" @click.self="showRegister = false">
      <div class="tb-modal" role="dialog" aria-modal="true" aria-label="도메인 등록">
        <h2 class="tb-modal-title">도메인 등록</h2>
        <div class="tb-field">
          <label class="tb-label" for="domain-name">도메인</label>
          <input id="domain-name" v-model="domainName" class="tb-input" placeholder="blog.example.com" />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="domain-type">연결 방식</label>
          <select id="domain-type" v-model="domainType" class="tb-select">
            <option value="CUSTOM_CNAME">CNAME (서브도메인 권장)</option>
            <option value="CUSTOM_APEX">Apex (A 레코드)</option>
          </select>
        </div>
        <div class="tb-modal-actions">
          <button class="tb-btn tb-btn-ghost" @click="showRegister = false">취소</button>
          <button class="tb-btn tb-btn-primary" :disabled="busy || !domainName.trim()" @click="register">
            {{ busy ? '등록 중…' : '등록' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="pending">
      <div v-for="index in 3" :key="index" class="tb-skeleton" style="height: 72px" />
    </div>
    <ul v-else-if="items.length > 0" class="tb-list">
      <li v-for="item in items" :key="item.id" class="tb-list-item">
        <div>
          <p class="tb-list-item-title">
            {{ item.domain_name }}
            <span v-if="item.is_primary" class="tb-badge">대표</span>
            <span class="tb-badge">{{ item.verification_status }}</span>
          </p>
          <p class="tb-list-item-desc">
            {{ item.domain_type }} · TLS {{ item.tls_status }}
            <span v-if="item.failure_reason"> · {{ item.failure_reason }}</span>
          </p>
          <code v-if="item.verification_status !== 'VERIFIED'" class="tb-code">
            TXT _treeblog-verify.{{ item.domain_name }} = {{ item.verification_token }}
          </code>
        </div>
        <div style="display: flex; gap: 6px; flex-wrap: wrap; justify-content: flex-end">
          <button
            v-if="item.verification_status !== 'VERIFIED'"
            class="tb-btn tb-btn-secondary tb-btn-sm"
            @click="verify(item.id)"
          >
            인증 확인
          </button>
          <button
            v-if="!item.is_primary && item.verification_status === 'VERIFIED'"
            class="tb-btn tb-btn-secondary tb-btn-sm"
            @click="setPrimary(item.id)"
          >
            대표로 설정
          </button>
          <button class="tb-btn tb-btn-ghost tb-btn-sm" @click="remove(item.id)">삭제</button>
        </div>
      </li>
    </ul>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">등록된 도메인이 없습니다</p>
      <p class="tb-empty-desc">treeblog.com/b/핸들 주소는 항상 사용할 수 있습니다.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, friendlyMessage } from '../lib/api'
import type { Domain } from '../lib/types'
import { useToastStore } from '../stores/toast'

const route = useRoute()
const toast = useToastStore()
const blogId = computed(() => String(route.params.blogId || ''))

const items = ref<Domain[]>([])
const pending = ref(true)
const busy = ref(false)
const showRegister = ref(false)
const domainName = ref('')
const domainType = ref('CUSTOM_CNAME')

async function load(): Promise<void> {
  pending.value = true
  try {
    items.value = await api<Domain[]>(`/api/v1/blogs/${blogId.value}/domains/`)
  } catch (err) {
    items.value = []
    toast.push(friendlyMessage(err), 'error')
  } finally {
    pending.value = false
  }
}

async function register(): Promise<void> {
  busy.value = true
  try {
    const created = await api<Domain>(`/api/v1/blogs/${blogId.value}/domains/`, {
      method: 'POST',
      body: { domain_name: domainName.value.trim(), domain_type: domainType.value }
    })
    items.value = [...items.value, created]
    showRegister.value = false
    domainName.value = ''
    toast.push('도메인을 등록했습니다. DNS TXT 레코드를 추가한 뒤 인증을 확인하세요', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    busy.value = false
  }
}

async function verify(domainId: string): Promise<void> {
  try {
    const updated = await api<Domain>(`/api/v1/blogs/${blogId.value}/domains/${domainId}/verify`, {
      method: 'POST',
      body: {}
    })
    const index = items.value.findIndex((item) => item.id === domainId)
    if (index >= 0) items.value[index] = updated
    toast.push(
      updated.verification_status === 'VERIFIED' ? '도메인이 인증되었습니다' : '아직 인증되지 않았습니다. DNS 전파를 기다리세요',
      updated.verification_status === 'VERIFIED' ? 'success' : 'info'
    )
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  }
}

async function setPrimary(domainId: string): Promise<void> {
  try {
    await api(`/api/v1/blogs/${blogId.value}/domains/${domainId}/primary`, {
      method: 'PUT',
      body: {}
    })
    items.value.forEach((item) => {
      item.is_primary = item.id === domainId
    })
    toast.push('대표 도메인을 변경했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  }
}

async function remove(domainId: string): Promise<void> {
  if (!confirm('이 도메인을 삭제할까요?')) return
  try {
    await api(`/api/v1/blogs/${blogId.value}/domains/${domainId}`, { method: 'DELETE' })
    items.value = items.value.filter((item) => item.id !== domainId)
    toast.push('삭제했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  }
}

onMounted(load)
</script>
