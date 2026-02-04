import { ref, onUnmounted } from 'vue'
import type { SSEEvent } from '@/types'

export interface UseSSEOptions {
  onOpen?: () => void
  onError?: (error: Error) => void
  onClose?: () => void
}

export function useSSE<T = any>(options: UseSSEOptions = {}) {
  const data = ref<T[]>([])
  const error = ref<Error | null>(null)
  const status = ref<'idle' | 'connecting' | 'connected' | 'error' | 'closed'>('idle')

  let abortFn: (() => void) | null = null

  function connect(url: string, body?: any) {
    // Abort previous connection
    disconnect()

    status.value = 'connecting'
    error.value = null
    data.value = []

    const controller = new AbortController()

    fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: body ? JSON.stringify(body) : undefined,
      signal: controller.signal
    })
      .then(async (response) => {
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`)
        }

        status.value = 'connected'
        options.onOpen?.()

        const reader = response.body?.getReader()
        if (!reader) {
          throw new Error('No response body')
        }

        const decoder = new TextDecoder()
        let buffer = ''

        try {
          while (true) {
            const { done, value } = await reader.read()
            if (done) break

            buffer += decoder.decode(value, { stream: true })
            const lines = buffer.split('\n')
            buffer = lines.pop() || ''

            for (const line of lines) {
              if (line.startsWith('data: ')) {
                try {
                  const eventData = JSON.parse(line.slice(6))
                  data.value.push(eventData as any)
                } catch (e) {
                  // Ignore parse errors
                }
              }
            }
          }
        } finally {
          reader.releaseLock()
          status.value = 'closed'
          options.onClose?.()
        }
      })
      .catch((err) => {
        if (err.name !== 'AbortError') {
          error.value = err
          status.value = 'error'
          options.onError?.(err)
        }
      })

    abortFn = () => controller.abort()
  }

  function disconnect() {
    if (abortFn) {
      abortFn()
      abortFn = null
    }
    status.value = 'idle'
  }

  function clear() {
    data.value = []
    error.value = null
  }

  // Auto cleanup
  onUnmounted(() => {
    disconnect()
  })

  return {
    data,
    error,
    status,
    connect,
    disconnect,
    clear
  }
}

/**
 * Process SSE events with typed handlers
 */
export function useSSEEvents<T extends string = string>() {
  const events = ref<SSEEvent<any>[]>([])
  const handlers = new Map<T, (data: any) => void>()

  function on(eventType: T, handler: (data: any) => void) {
    handlers.set(eventType, handler)
  }

  function off(eventType: T) {
    handlers.delete(eventType)
  }

  function handleEvent(event: SSEEvent) {
    events.value.push(event)
    const handler = handlers.get(event.type as T)
    if (handler) {
      handler(event.data)
    }
  }

  function clear() {
    events.value = []
  }

  return {
    events,
    on,
    off,
    handleEvent,
    clear
  }
}
