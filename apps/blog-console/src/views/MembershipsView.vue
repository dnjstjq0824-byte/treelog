<template>
  <div>
    <div class="tb-console-head">
      <h1 class="tb-console-title">멤버십 요금제</h1>
      <button class="tb-btn tb-btn-primary tb-btn-sm" @click="showCreate = true">새 요금제</button>
    </div>

    <div v-if="showCreate" class="tb-modal-backdrop" @click.self="showCreate = false">
      <div class="tb-modal" role="dialog" aria-modal="true" aria-label="새 요금제">
        <h2 class="tb-modal-title">새 요금제</h2>
        <div class="tb-field">
          <label class="tb-label" for="tier-name">이름</label>
          <input id="tier-name" v-model="form.name" class="tb-input" required maxlength="60" />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="tier-desc">설명</label>
          <input id="tier-desc" v-model="form.description" class="tb-input" maxlength="300" />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="tier-monthly">월 가격 (원)</label>
          <input id="tier-monthly" v-model.number="form.monthly_price_minor" class="tb-input" type="number" min="0" step="100" />
        </div>
        <div class="tb-field">
          <label class="tb-label" for="tier-annual">연 가격 (원, 선택)</label>
          <input id="tier-annual" v-model.number="form.annual_price_minor" class="tb-input" type="number" min="0" step="100" placeholder="미제공" />
        </div>
        <div class="tb-modal-actions">
          <button class="tb-btn tb-btn-ghost" @click="showCreate = false">취소</button>
          <button class="tb-btn tb-btn-primary" :disabled="busy || !form.name.trim()" @click="create">
            {{ busy ? '생성 중…' : '생성' }}
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
          <p class="tb-list-item-title">{{ item.name }}</p>
          <p class="tb-list-item-desc">
            {{ item.description }} · 월 {{ item.monthly_price_minor.toLocaleString() }}원
            <template v-if="item.annual_price_minor"> · 연 {{ item.annual_price_minor.toLocaleString() }}원</template>
          </p>
        </div>
        <button class="tb-btn tb-btn-ghost tb-btn-sm" @click="deactivate(item.id)">판매 중지</button>
      </li>
    </ul>
    <div v-else class="tb-empty">
      <p class="tb-empty-title">요금제가 없습니다</p>
      <p class="tb-empty-desc">유료 멤버십을 판매하려면 요금제를 만드세요.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { api, friendlyMessage } from '../lib/api'
import type { MembershipTier } from '../lib/types'
import { useToastStore } from '../stores/toast'

const route = useRoute()
const toast = useToastStore()
const blogId = computed(() => String(route.params.blogId || ''))

const items = ref<MembershipTier[]>([])
const pending = ref(true)
const busy = ref(false)
const showCreate = ref(false)
const form = ref({ name: '', description: '', monthly_price_minor: 4900, annual_price_minor: 0 })

async function load(): Promise<void> {
  pending.value = true
  try {
    items.value = await api<MembershipTier[]>(`/api/v1/blogs/${blogId.value}/memberships/`)
  } catch (err) {
    items.value = []
    toast.push(friendlyMessage(err), 'error')
  } finally {
    pending.value = false
  }
}

async function create(): Promise<void> {
  busy.value = true
  try {
    const created = await api<MembershipTier>(`/api/v1/blogs/${blogId.value}/memberships/`, {
      method: 'POST',
      body: {
        name: form.value.name.trim(),
        description: form.value.description.trim(),
        monthly_price_minor: form.value.monthly_price_minor,
        annual_price_minor: form.value.annual_price_minor > 0 ? form.value.annual_price_minor : null,
        benefits_json: '[]'
      }
    })
    items.value = [...items.value, created]
    showCreate.value = false
    form.value = { name: '', description: '', monthly_price_minor: 4900, annual_price_minor: 0 }
    toast.push('요금제를 만들었습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  } finally {
    busy.value = false
  }
}

async function deactivate(membershipId: string): Promise<void> {
  if (!confirm('이 요금제의 판매를 중지할까요? 기존 구독은 유지됩니다.')) return
  try {
    await api(`/api/v1/blogs/${blogId.value}/memberships/${membershipId}`, { method: 'DELETE' })
    items.value = items.value.filter((item) => item.id !== membershipId)
    toast.push('판매를 중지했습니다', 'success')
  } catch (err) {
    toast.push(friendlyMessage(err), 'error')
  }
}

onMounted(load)
</script>
