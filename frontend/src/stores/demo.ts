import { defineStore } from 'pinia'
import { ref } from 'vue'
import type {
  GroundingResult,
  ReActStep,
  RichContext,
  ComparisonCase,
  ComparisonResult,
  MaintenanceLog
} from '@/types'
import { comparisonApi, selfMaintainApi, queryApi } from '@/api'

export const useDemoStore = defineStore('demo', () => {
  // Current scenario tab
  const currentScenario = ref<'benchmark' | 'comparison' | 'selfmaintain'>('comparison')

  // Query state
  const question = ref('')
  const databaseId = ref('spider_sqlite')
  const database = ref('')
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Options
  const options = ref({
    useRichContext: true,
    useReact: true,
    useGrounding: true,
    maxIterations: 5
  })

  // Results
  const sql = ref('')
  const groundingResult = ref<GroundingResult | null>(null)
  const reactSteps = ref<ReActStep[]>([])
  const usedContexts = ref<RichContext[]>([])
  const duration = ref(0)

  // Grounding state
  const groundingStage = ref<'idle' | 'stage1' | 'stage2' | 'done'>('idle')
  const groundingDuration = ref({ stage1: 0, stage2: 0 })

  // Comparison state
  const comparisonCases = ref<ComparisonCase[]>([])
  const selectedCase = ref<ComparisonCase | null>(null)
  const comparisonResult = ref<ComparisonResult | null>(null)
  const isComparing = ref(false)

  // Self-maintain state
  const maintenanceLogs = ref<MaintenanceLog[]>([])

  // Abort controller
  let abortFn: (() => void) | null = null

  // Actions
  function reset() {
    sql.value = ''
    groundingResult.value = null
    reactSteps.value = []
    usedContexts.value = []
    duration.value = 0
    groundingStage.value = 'idle'
    groundingDuration.value = { stage1: 0, stage2: 0 }
    error.value = null
    isLoading.value = false
  }

  function abort() {
    if (abortFn) {
      abortFn()
      abortFn = null
    }
    isLoading.value = false
  }

  async function runQuery() {
    reset()
    isLoading.value = true
    const startTime = Date.now()

    abortFn = queryApi.stream(
      {
        question: question.value,
        databaseId: databaseId.value,
        database: database.value,
        options: options.value
      },
      (event: { type: string; data: any }) => {
        switch (event.type) {
          case 'grounding_start':
            groundingStage.value = 'stage1'
            break
          case 'grounding_stage1':
            groundingStage.value = 'stage1'
            groundingDuration.value.stage1 = event.data.duration || 0
            break
          case 'grounding_stage2':
            groundingStage.value = 'stage2'
            groundingDuration.value.stage2 = event.data.duration || 0
            break
          case 'grounding_complete':
            groundingStage.value = 'done'
            groundingResult.value = event.data
            break
          case 'react_step':
            reactSteps.value.push(event.data)
            break
          case 'complete':
            sql.value = event.data.sql || event.data.final_sql || ''
            duration.value = Date.now() - startTime
            isLoading.value = false
            if (event.data.used_contexts) {
              usedContexts.value = event.data.used_contexts
            }
            break
          case 'error':
            error.value = event.data.message || 'Unknown error'
            isLoading.value = false
            break
        }
      },
      (err: Error) => {
        error.value = err.message
        isLoading.value = false
      }
    )
  }

  async function loadComparisonCases() {
    comparisonCases.value = await comparisonApi.getCases()
  }

  async function runComparison(caseItem: ComparisonCase) {
    selectedCase.value = caseItem
    isComparing.value = true
    comparisonResult.value = null

    try {
      comparisonResult.value = await comparisonApi.runComparison(caseItem.id)
    } catch (e: any) {
      error.value = e.message
    } finally {
      isComparing.value = false
    }
  }

  async function loadMaintenanceLogs() {
    maintenanceLogs.value = await selfMaintainApi.getLogs()
  }

  async function triggerMaintenance(type: string) {
    try {
      const log = await selfMaintainApi.triggerMaintenance(type)
      maintenanceLogs.value.unshift(log)
    } catch (e: any) {
      error.value = e.message
    }
  }

  return {
    // State
    currentScenario,
    question,
    databaseId,
    database,
    isLoading,
    error,
    options,
    sql,
    groundingResult,
    groundingStage,
    groundingDuration,
    reactSteps,
    usedContexts,
    duration,
    comparisonCases,
    selectedCase,
    comparisonResult,
    isComparing,
    maintenanceLogs,

    // Actions
    reset,
    abort,
    runQuery,
    loadComparisonCases,
    runComparison,
    loadMaintenanceLogs,
    triggerMaintenance
  }
})
