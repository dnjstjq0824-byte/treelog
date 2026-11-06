import { defineStore } from 'pinia'

export interface Toast {
  id: number
  message: string
  kind: 'info' | 'success' | 'error'
}

let nextId = 1

export const useToastStore = defineStore('toast', {
  state: () => ({
    toasts: [] as Toast[]
  }),
  actions: {
    push(message: string, kind: Toast['kind'] = 'info'): void {
      const id = nextId++
      this.toasts.push({ id, message, kind })
      setTimeout(() => {
        this.toasts = this.toasts.filter((toast) => toast.id !== id)
      }, 3600)
    }
  }
})
