import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // Landing Page - Database Collection
    {
      path: '/',
      name: 'Landing',
      component: () => import('@/features/landing/index.vue'),
      meta: { title: 'LUCID - My Databases' }
    },

    // Workspace - Database specific
    {
      path: '/workspace/:databaseId',
      name: 'Workspace',
      component: () => import('@/features/workspace/index.vue'),
      meta: { title: '工作区' }
    },

    // Demo Showcase
    {
      path: '/demo',
      name: 'Demo',
      component: () => import('@/features/demo/index.vue'),
      meta: { title: 'Demo 展示' }
    },

    // Settings (placeholder)
    {
      path: '/settings',
      name: 'Settings',
      component: () => import('@/features/landing/index.vue'), // Placeholder
      meta: { title: '设置' }
    },

    // Catch all - redirect to landing
    {
      path: '/:pathMatch(.*)*',
      redirect: '/'
    }
  ]
})

// Navigation guards
router.beforeEach((to, from, next) => {
  // Update document title
  document.title = (to.meta.title as string) || 'LUCID'
  next()
})

export default router
