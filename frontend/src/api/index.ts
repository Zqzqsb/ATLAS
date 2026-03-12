// Re-export all API modules
export { default as client } from './client'
export { createSSEStream } from './client'
export { databaseApi } from './database'
export { contextApi } from './context'
export { queryApi } from './query'
export { agentApi } from './agent'
export type { ChangeLog } from './agent'
export { evolutionApi } from './evolution'
export type { EvolutionStatus, EvolutionStage, StageExecution, ContextAction, EvolutionEvent } from './evolution'


