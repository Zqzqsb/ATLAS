import { request, createSSEStream } from './client'

// ========================================
// Types
// ========================================

export interface EvolutionStage {
  id: number
  name: string
  description: string
  ddls: string[]
  sample_data?: string[]
  expected_changes: string[]
  executed: boolean
  is_next: boolean
}

export interface StageExecution {
  stage_id: number
  stage_name: string
  ddl_executed: string[]
  changes_detected: SchemaChange[]
  context_actions: ContextAction[]
  executed_at: string
  duration_ms: number
  success: boolean
  error?: string
}

export interface SchemaChange {
  change_type: string
  table_name: string
  column_name?: string
  old_definition?: string
  new_definition?: string
  details?: Record<string, any>
  detected_at: string
}

export interface ContextAction {
  action_type: string  // "created" | "expired" | "refreshed" | "deleted"
  table_name: string
  column_name?: string
  context_type?: string
  description: string
  old_content?: string
  new_content?: string
}

export interface EvolutionStatus {
  current_stage: number
  total_stages: number
  database_name: string
  is_ready: boolean
  stages: EvolutionStage[]
  history: StageExecution[]
}

export interface EvolutionEvent {
  type: string
  phase: string
  message: string
  data?: any
}

// ========================================
// API
// ========================================

export const evolutionApi = {
  /**
   * Get current evolution status
   */
  getStatus: () =>
    request<EvolutionStatus>({ url: '/evolution/status', method: 'GET' }),

  /**
   * Preview a specific stage
   */
  getStagePreview: (stageId: number) =>
    request<EvolutionStage>({ url: `/evolution/stages/${stageId}`, method: 'GET' }),

  /**
   * Execute a stage (non-streaming)
   */
  executeStage: (datasourceId: number, stage: number) =>
    request<{ success: boolean; execution: StageExecution }>({
      url: '/evolution/execute-stage',
      method: 'POST',
      data: { datasource_id: datasourceId, stage }
    }),

  /**
   * Execute a stage with SSE streaming for real-time events
   */
  executeStageStream: (
    datasourceId: number,
    stage: number,
    onEvent: (event: { type: string; data: any }) => void,
    onError?: (error: Error) => void,
    onComplete?: () => void
  ) => {
    return createSSEStream(
      '/api/v1/evolution/execute-stage/stream',
      { datasource_id: datasourceId, stage },
      onEvent,
      onError,
      onComplete
    )
  },

  /**
   * Reset to initial state
   */
  reset: (datasourceId: number) =>
    request<{ success: boolean; message: string; current_stage: number }>({
      url: '/evolution/reset',
      method: 'POST',
      data: { datasource_id: datasourceId }
    }),

  /**
   * Reset with SSE streaming
   */
  resetStream: (
    datasourceId: number,
    onEvent: (event: { type: string; data: any }) => void,
    onError?: (error: Error) => void,
    onComplete?: () => void
  ) => {
    return createSSEStream(
      '/api/v1/evolution/reset/stream',
      { datasource_id: datasourceId },
      onEvent,
      onError,
      onComplete
    )
  }
}
