import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Toast } from '@/types'

export const useAppStore = defineStore('app', () => {
  // Theme
  const isDarkMode = ref(false) // Always false
  const locale = ref<'zh' | 'en'>('zh')

  // UI State
  const sidebarCollapsed = ref(false)
  const toasts = ref<Toast[]>([])

  // Global loading
  const globalLoading = ref(false)
  const globalLoadingText = ref('')

  // Computed
  const theme = computed(() => 'light')

  // Actions
  function toggleDarkMode() {
    // Disabled
    isDarkMode.value = false
    document.documentElement.classList.remove('dark')
  }

  function setLocale(newLocale: 'zh' | 'en') {
    locale.value = newLocale
  }

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function showToast(toast: Omit<Toast, 'id'>) {
    const id = `toast-${Date.now()}`
    const newToast: Toast = { ...toast, id }
    toasts.value.push(newToast)

    // Auto remove after duration
    const duration = toast.duration ?? 3000
    if (duration > 0) {
      setTimeout(() => {
        removeToast(id)
      }, duration)
    }

    return id
  }

  function removeToast(id: string) {
    const index = toasts.value.findIndex(t => t.id === id)
    if (index >= 0) {
      toasts.value.splice(index, 1)
    }
  }

  function showSuccess(title: string, message?: string) {
    return showToast({ type: 'success', title, message })
  }

  function showError(title: string, message?: string) {
    return showToast({ type: 'error', title, message, duration: 5000 })
  }

  function showWarning(title: string, message?: string) {
    return showToast({ type: 'warning', title, message })
  }

  function showInfo(title: string, message?: string) {
    return showToast({ type: 'info', title, message })
  }

  function setGlobalLoading(loading: boolean, text = '') {
    globalLoading.value = loading
    globalLoadingText.value = text
  }

  return {
    // State
    isDarkMode,
    locale,
    sidebarCollapsed,
    toasts,
    globalLoading,
    globalLoadingText,

    // Computed
    theme,

    // Actions
    toggleDarkMode,
    setLocale,
    toggleSidebar,
    showToast,
    removeToast,
    showSuccess,
    showError,
    showWarning,
    showInfo,
    setGlobalLoading
  }
})
