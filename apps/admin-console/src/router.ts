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
        { path: '', name: 'dashboard', component: () => import('./views/DashboardView.vue') },
        { path: 'users', name: 'users', component: () => import('./views/UsersView.vue') },
        { path: 'blogs', name: 'blogs', component: () => import('./views/BlogsView.vue') },
        { path: 'moderation', name: 'cases', component: () => import('./views/CasesView.vue') },
        { path: 'moderation/:caseId', name: 'case-detail', component: () => import('./views/CaseDetailView.vue') },
        { path: 'appeals', name: 'appeals', component: () => import('./views/AppealsView.vue') },
        { path: 'payouts', name: 'payouts', component: () => import('./views/PayoutsView.vue') }
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
