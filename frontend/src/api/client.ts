import axios from 'axios'
import type { AxiosInstance, AxiosError, AxiosRequestConfig } from 'axios'
import type { ApiResponse } from '@/types'

// Create axios instance
const client: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 60000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor
client.interceptors.request.use(
  (config) => {
    // Add auth token if exists
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor
client.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    // Handle common errors
    if (error.response) {
      switch (error.response.status) {
        case 401:
          // Handle unauthorized
          console.error('Unauthorized')
          break
        case 403:
          console.error('Forbidden')
          break
        case 500:
          console.error('Server error')
          break
      }
    }
    return Promise.reject(error)
  }
)

// Generic request wrapper
export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const response = await client.request<T>(config)
  return response.data
}

// SSE Stream helper - handles standard SSE format with event: and data: lines
export function createSSEStream<T>(
  url: string,
  body: any,
  onEvent: (event: { type: string; data: T }) => void,
  onError?: (error: Error) => void,
  onComplete?: () => void
): () => void {
  const controller = new AbortController()

  fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
    signal: controller.signal
  })
    .then(async (response) => {
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const reader = response.body?.getReader()
      if (!reader) {
        throw new Error('No response body')
      }

      const decoder = new TextDecoder()
      let buffer = ''
      let currentEventType = 'message' // default SSE event type

      // Helper: yield control to the browser so Vue can render between events
      const yieldToRenderer = () => new Promise<void>(resolve => setTimeout(resolve, 0))

      try {
        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          buffer += decoder.decode(value, { stream: true })
          const lines = buffer.split('\n')
          buffer = lines.pop() || ''

          // Collect events from this chunk, then dispatch with yields between them
          const pendingEvents: { type: string; data: T }[] = []

          for (const line of lines) {
            // Parse event type line
            if (line.startsWith('event: ')) {
              currentEventType = line.slice(7).trim()
            }
            // Parse data line
            else if (line.startsWith('data: ')) {
              try {
                const data = JSON.parse(line.slice(6))
                pendingEvents.push({ type: currentEventType, data: data as T })
                // Reset event type after collecting
                currentEventType = 'message'
              } catch (e) {
                // ignore parse errors
              }
            }
            // Empty line marks end of event block (SSE spec)
            else if (line === '') {
              currentEventType = 'message'
            }
          }

          // Dispatch events; if multiple events in one chunk, yield between them
          // so Vue can render intermediate states (progressive SSE)
          for (let i = 0; i < pendingEvents.length; i++) {
            onEvent(pendingEvents[i])
            if (i < pendingEvents.length - 1) {
              await yieldToRenderer()
            }
          }
        }
      } finally {
        reader.releaseLock()
        onComplete?.()
      }
    })
    .catch((error) => {
      if (error.name !== 'AbortError') {
        onError?.(error)
      }
    })

  return () => controller.abort()
}

// Named export for use in API modules
export const apiClient = client

export default client
