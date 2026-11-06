import { defineStore } from 'pinia'
import { api, clearAccess, saveAccess } from '../lib/api'
import type { SessionUser } from '../lib/types'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as SessionUser | null,
    initialized: false
  }),
  getters: {
    isLoggedIn: (state) => state.user !== null
  },
  actions: {
    async fetchMe(): Promise<SessionUser | null> {
      try {
        this.user = await api<SessionUser>('/api/v1/me/')
        return this.user
      } catch {
        this.user = null
        return null
      }
    },
    async initialize(): Promise<void> {
      if (this.initialized) return
      this.initialized = true
      if (localStorage.getItem('tb_access')) await this.fetchMe()
    },
    async login(email: string, password: string): Promise<void> {
      const tokens = await api<{ access_token: string; expires_in: number }>('/api/v1/auth/login', {
        method: 'POST',
        auth: false,
        body: { email, password }
      })
      saveAccess(tokens.access_token)
      await this.fetchMe()
    },
    async logout(): Promise<void> {
      try {
        await api('/api/v1/auth/logout', { method: 'POST' })
      } catch {
        // Clear locally regardless of server state.
      }
      this.user = null
      clearAccess()
    }
  }
})
