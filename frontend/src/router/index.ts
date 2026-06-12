import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // Landing Page - Database Collection
    {
      path: '/',
      name: 'Landing',
      component: () => import('@/features/landing/index.vue'),
      meta: { title: 'ATLAS - My Databases' }
    },

    // Feature Showcase
    {
      path: '/features',
      name: 'Features',
      component: () => import('@/features/showcase/index.vue'),
      meta: { title: 'ATLAS - Feature Showcase' }
    },

    // Workspace - Database specific
    {
      path: '/workspace/:databaseId',
      name: 'Workspace',
      component: () => import('@/features/workspace/index.vue'),
      meta: { title: 'Workspace' }
    },

    // Architecture Diagram
    {
      path: '/arch',
      name: 'Arch',
      component: () => import('@/features/arch/index.vue'),
      meta: { title: 'ATLAS - Architecture' }
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
  document.title = (to.meta.title as string) || 'ATLAS'
  next()
})

export default router
