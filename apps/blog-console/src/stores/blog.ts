import { defineStore } from 'pinia'
import { api } from '../lib/api'
import type { Blog } from '../lib/types'

export const useBlogStore = defineStore('blog', {
  state: () => ({
    blogs: [] as Blog[],
    currentId: localStorage.getItem('tb_blog_id') || '',
    loaded: false
  }),
  getters: {
    current(state): Blog | null {
      return state.blogs.find((blog) => blog.id === state.currentId) || null
    }
  },
  actions: {
    async load(): Promise<void> {
      this.blogs = await api<Blog[]>('/api/v1/blogs/my')
      if (!this.blogs.some((blog) => blog.id === this.currentId)) {
        this.currentId = this.blogs[0]?.id || ''
        localStorage.setItem('tb_blog_id', this.currentId)
      }
      this.loaded = true
    },
    select(id: string): void {
      this.currentId = id
      localStorage.setItem('tb_blog_id', id)
    }
  }
})
