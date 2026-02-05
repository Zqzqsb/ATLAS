import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface WarmupStatus {
  warmed: boolean
  warmup_duration: string
  schema_cache_size: number
}

export interface SystemHealth {
  status: string
  timestamp: number
  warmup: WarmupStatus
}

export const useSystemStore = defineStore('system', () => {
  // State
  const health = ref<SystemHealth | null>(null)
  const loading = ref(false)
  const lastChecked = ref<Date | null>(null)
  const checkInterval = ref<number | null>(null)

  // Computed
  const isWarmedUp = computed(() => health.value?.warmup?.warmed ?? false)
  const isHealthy = computed(() => health.value?.status === 'ok')
  const warmupDuration = computed(() => health.value?.warmup?.warmup_duration ?? '')

  // Actions
  async function checkHealth(): Promise<SystemHealth | null> {
    loading.value = true
    try {
      const response = await fetch('/health')
      if (!response.ok) {
        throw new Error(`Health check failed: ${response.status}`)
      }
      health.value = await response.json()
      lastChecked.value = new Date()
      return health.value
    } catch (e) {
      console.error('Health check failed:', e)
      return null
    } finally {
      loading.value = false
    }
  }

  // Start periodic health checks
  function startHealthCheck(intervalMs: number = 10000) {
    if (checkInterval.value) {
      clearInterval(checkInterval.value)
    }
    
    // Initial check
    checkHealth()
    
    // Periodic checks
    checkInterval.value = window.setInterval(() => {
      checkHealth()
    }, intervalMs)
  }

  // Stop periodic health checks
  function stopHealthCheck() {
    if (checkInterval.value) {
      clearInterval(checkInterval.value)
      checkInterval.value = null
    }
  }

  // Warmup status helpers
  function waitForWarmup(timeoutMs: number = 30000): Promise<boolean> {
    return new Promise((resolve) => {
      const startTime = Date.now()
      
      const check = async () => {
        await checkHealth()
        
        if (isWarmedUp.value) {
          resolve(true)
          return
        }
        
        if (Date.now() - startTime > timeoutMs) {
          resolve(false)
          return
        }
        
        // Check again in 1 second
        setTimeout(check, 1000)
      }
      
      check()
    })
  }

  return {
    // State
    health,
    loading,
    lastChecked,
    
    // Computed
    isWarmedUp,
    isHealthy,
    warmupDuration,
    
    // Actions
    checkHealth,
    startHealthCheck,
    stopHealthCheck,
    waitForWarmup
  }
})
