import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from './stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: () => import('./views/LoginView.vue') },
    {
      path: '/',
      component: () => import('./views/ShellView.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', name: 'blogs', component: () => import('./views/BlogSelectView.vue') },
        { path: 'b/:blogId/posts', name: 'posts', component: () => import('./views/PostsView.vue') },
        { path: 'b/:blogId/posts/new', name: 'post-new', component: () => import('./views/EditorView.vue') },
        { path: 'b/:blogId/posts/:postId', name: 'post-edit', component: () => import('./views/EditorView.vue') },
        { path: 'b/:blogId/media', name: 'media', component: () => import('./views/MediaView.vue') },
        { path: 'b/:blogId/categories', name: 'categories', component: () => import('./views/CategoriesView.vue') },
        { path: 'b/:blogId/series', name: 'series', component: () => import('./views/SeriesView.vue') },
        { path: 'b/:blogId/guestbook', name: 'guestbook', component: () => import('./views/GuestbookView.vue') },
        { path: 'b/:blogId/members', name: 'members', component: () => import('./views/MembersView.vue') },
        { path: 'b/:blogId/settings', name: 'settings', component: () => import('./views/SettingsView.vue') },
        { path: 'b/:blogId/theme', name: 'theme', component: () => import('./views/ThemeView.vue') },
        { path: 'b/:blogId/domains', name: 'domains', component: () => import('./views/DomainsView.vue') },
        { path: 'b/:blogId/memberships', name: 'memberships', component: () => import('./views/MembershipsView.vue') },
        { path: 'b/:blogId/payout', name: 'payout', component: () => import('./views/PayoutView.vue') }
      ]
    },
    { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('./views/NotFoundView.vue') }
  ]
})

router.beforeEach(async (to) => {
  if (!to.meta.requiresAuth) return true
  const auth = useAuthStore()
  await auth.initialize()
  if (!auth.isLoggedIn) return { name: 'login', query: { next: to.fullPath } }
  return true
})

export default router
