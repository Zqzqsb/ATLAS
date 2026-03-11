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

    // Feature Showcase
    {
      path: '/features',
      name: 'Features',
      component: () => import('@/features/showcase/index.vue'),
      meta: { title: 'LUCID - Feature Showcase' }
    },

    // Workspace - Database specific
    {
      path: '/workspace/:databaseId',
      name: 'Workspace',
      component: () => import('@/features/workspace/index.vue'),
      meta: { title: 'Workspace' }
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
