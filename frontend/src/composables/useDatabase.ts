import { computed } from 'vue'
import { useDatabaseStore } from '@/stores/database'
import { useWorkspaceStore } from '@/stores/workspace'
import type { Database } from '@/types'

/**
 * Composable for database operations
 */
export function useDatabase() {
  const databaseStore = useDatabaseStore()
  const workspaceStore = useWorkspaceStore()

  // Computed
  const databases = computed(() => databaseStore.databases)
  const currentDatabase = computed(() => workspaceStore.currentDatabase)
  const currentDatabaseId = computed(() => workspaceStore.currentDatabaseId)
  const isConnected = computed(() => currentDatabase.value?.status === 'connected')
  const hasRichContext = computed(() => currentDatabase.value?.hasRichContext ?? false)

  // Methods
  async function selectDatabase(id: string) {
    await workspaceStore.selectDatabase(id)
  }

  function getDatabaseById(id: string): Database | undefined {
    return databaseStore.getDatabaseById(id)
  }

  async function refreshDatabases() {
    await databaseStore.fetchDatabases()
  }

  async function testConnection(id: string) {
    return await databaseStore.testConnection(id)
  }

  return {
    // State
    databases,
    currentDatabase,
    currentDatabaseId,
    isConnected,
    hasRichContext,

    // Methods
    selectDatabase,
    getDatabaseById,
    refreshDatabases,
    testConnection
  }
}

/**
 * Composable for clipboard operations
 */
export function useClipboard() {
  async function copy(text: string): Promise<boolean> {
    try {
      await navigator.clipboard.writeText(text)
      return true
    } catch {
      // Fallback for older browsers
      const textarea = document.createElement('textarea')
      textarea.value = text
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      try {
        document.execCommand('copy')
        return true
      } catch {
        return false
      } finally {
        document.body.removeChild(textarea)
      }
    }
  }

  async function paste(): Promise<string | null> {
    try {
      return await navigator.clipboard.readText()
    } catch {
      return null
    }
  }

  return { copy, paste }
}

/**
 * Composable for responsive breakpoints
 */
export function useBreakpoint() {
  const getBreakpoint = () => {
    const width = window.innerWidth
    if (width < 640) return 'xs'
    if (width < 768) return 'sm'
    if (width < 1024) return 'md'
    if (width < 1280) return 'lg'
    if (width < 1536) return 'xl'
    return '2xl'
  }

  const isMobile = computed(() => {
    const bp = getBreakpoint()
    return bp === 'xs' || bp === 'sm'
  })

  const isTablet = computed(() => {
    const bp = getBreakpoint()
    return bp === 'md'
  })

  const isDesktop = computed(() => {
    const bp = getBreakpoint()
    return bp === 'lg' || bp === 'xl' || bp === '2xl'
  })

  return {
    getBreakpoint,
    isMobile,
    isTablet,
    isDesktop
  }
}
