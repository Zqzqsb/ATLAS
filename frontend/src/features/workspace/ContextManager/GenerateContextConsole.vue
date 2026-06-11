<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { NModal, NButton, NInputNumber, NProgress, NSwitch, useMessage } from 'naive-ui'
import { useContextGenerationStore, type ChunkClusterInfo, type ForestClusterPreview } from '@/stores/contextGeneration'

const props = defineProps<{
  show: boolean
  databaseId: string
  tableCount?: number
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'complete'): void
  (e: 'minimize'): void
}>()

const store = useContextGenerationStore()
const message = useMessage()

// Expose state for parent component
defineExpose({
  isRunning: computed(() => store.isRunning),
  isComplete: computed(() => store.isComplete),
  progress: computed(() => store.overallProgress)
})

// Computed
const showModal = computed({
  get: () => props.show,
  set: (v) => emit('update:show', v)
})

// Whether forest mode is active (for cancel button logic)
const isForestRunning = computed(() => store.isRunning && store.chunkProgress.isForestMode)

// ----- Treemap helpers -----

interface TreemapRect {
  cluster: ChunkClusterInfo
  x: number
  y: number
  w: number
  h: number
}

/** Gap in px between treemap cells */
const GAP = 2

/**
 * Squarified treemap layout (Bruls–Huizing–van Wijk) with inset gaps.
 *
 * The algorithm works on the full container area for correct proportions,
 * then insets each rect by GAP/2 on every side so cells don't touch.
 */
function squarify(items: ChunkClusterInfo[], containerW: number, containerH: number): TreemapRect[] {
  if (items.length === 0 || containerW <= 0 || containerH <= 0) return []

  const totalArea = containerW * containerH
  const totalValue = items.reduce((s, c) => s + Math.max(c.tableCount, 1), 0)
  if (totalValue === 0) return []

  const areas: number[] = items.map(c => (Math.max(c.tableCount, 1) / totalValue) * totalArea)

  const raw: { idx: number; x: number; y: number; w: number; h: number }[] = []

  function getArea(idx: number): number { return areas[idx] ?? 0 }

  function worst(row: number[], side: number): number {
    if (row.length === 0 || side <= 0) return Infinity
    let sum = 0, lo = Infinity, hi = -Infinity
    for (const idx of row) {
      const a = getArea(idx)
      sum += a
      if (a < lo) lo = a
      if (a > hi) hi = a
    }
    if (sum === 0) return Infinity
    const s2 = side * side, r2 = sum * sum
    return Math.max((s2 * hi) / r2, r2 / (s2 * lo))
  }

  const order: number[] = Array.from({ length: items.length }, (_, i) => i)
    .sort((a, b) => getArea(b) - getArea(a))

  function recurse(rem: number[], x: number, y: number, w: number, h: number) {
    if (rem.length === 0) return
    if (rem.length === 1) { raw.push({ idx: rem[0]!, x, y, w, h }); return }

    // Always slice along the longest edge
    const isWide = w >= h
    const side = isWide ? h : w

    const first = rem[0]!
    const row: number[] = [first]
    let rowArea = getArea(first)
    let best = worst(row, side)

    let i = 1
    for (; i < rem.length; i++) {
      const c = rem[i]!
      const nr = [...row, c]
      const nw = worst(nr, side)
      if (nw <= best) {
        row.push(c)
        rowArea += getArea(c)
        best = nw
      } else {
        break
      }
    }

    if (isWide) {
      // Wide remaining space: layout a vertical column on the left edge
      const colW = rowArea / h
      let cy = y
      for (const idx of row) {
        const itemA = getArea(idx)
        const itemH = itemA / colW
        raw.push({ idx, x, y: cy, w: colW, h: itemH })
        cy += itemH
      }
      recurse(rem.slice(i), x + colW, y, Math.max(w - colW, 0), h)
    } else {
      // Tall remaining space: layout a horizontal row on the top edge
      const rowH = rowArea / w
      let cx = x
      for (const idx of row) {
        const itemA = getArea(idx)
        const itemW = itemA / rowH
        raw.push({ idx, x: cx, y, w: itemW, h: rowH })
        cx += itemW
      }
      recurse(rem.slice(i), x, y + rowH, w, Math.max(h - rowH, 0))
    }
  }

  recurse(order, 0, 0, containerW, containerH)

  // Apply gaps (inset each rect)
  const half = GAP / 2
  return raw.map(r => ({
    cluster: items[r.idx]!,
    x: r.x + half,
    y: r.y + half,
    w: Math.max(r.w - GAP, 1),
    h: Math.max(r.h - GAP, 1),
  }))
}

// Make sure width fits safely within 850px modal minus paddings
const TREEMAP_W = 760
const TREEMAP_H = 210

const treemapRects = computed(() => {
  const clusters = store.chunkProgress.clusters
  if (!clusters || clusters.length === 0) return []
  return squarify(clusters, TREEMAP_W, TREEMAP_H)
})

function cellBg(status: string): string {
  switch (status) {
    case 'success': return '#86efac'  // green-300
    case 'skipped': return '#bae6fd'  // sky-200 (lighter blue — already done)
    case 'running': return '#fcd34d'  // amber-300
    case 'error':   return '#fca5a5'  // red-300
    default:        return '#cbd5e1'  // slate-300 (pending)
  }
}

function cellBorder(status: string): string {
  switch (status) {
    case 'success': return '#22c55e'  // green-500
    case 'skipped': return '#38bdf8'  // sky-400
    case 'running': return '#f59e0b'  // amber-500
    case 'error':   return '#ef4444'  // red-500
    default:        return '#94a3b8'  // slate-400
  }
}

function cellText(status: string): string {
  switch (status) {
    case 'success': return '#14532d'  // green-900
    case 'skipped': return '#0c4a6e'  // sky-900
    case 'running': return '#78350f'  // amber-900
    case 'error':   return '#7f1d1d'  // red-900
    default:        return '#0f172a'  // slate-900
  }
}

/** Smart label for treemap cells — prioritize cluster index (sequence number) */
function cellLabel(rect: TreemapRect): string {
  const { cluster, w, h } = rect
  const idx = `#${cluster.index + 1}`
  // Large cell: show "#N · Xt"
  if (w >= 50 && h >= 24) return `${idx} · ${cluster.tableCount}t`
  // Medium cell: show "#N"
  if (w >= 22 && h >= 16) return idx
  return ''
}

// Tooltip state
const hoveredCluster = ref<ChunkClusterInfo | null>(null)
const tooltipPos = ref({ x: 0, y: 0 })

function onCellEnter(e: MouseEvent, cluster: ChunkClusterInfo) {
  hoveredCluster.value = cluster
  tooltipPos.value = { x: e.clientX, y: e.clientY }
}
function onCellMove(e: MouseEvent) {
  tooltipPos.value = { x: e.clientX, y: e.clientY }
}
function onCellLeave() {
  hoveredCluster.value = null
}

// Helper functions (purely presentational)
function getPhaseIcon(phase: string): string {
  switch (phase) {
    case 'thought': return '💭'
    case 'action': return '🔧'
    case 'observation': return '📊'
    case 'storage': return '💾'
    case 'success': return '✅'
    case 'error': return '❌'
    case 'finish': return '🏁'
    default: return '📝'
  }
}

function getPhaseColor(phase: string): string {
  switch (phase) {
    case 'thought': return 'text-gray-200'
    case 'action': return 'text-blue-300'
    case 'observation': return 'text-cyan-200'
    case 'storage': return 'text-emerald-300'
    case 'success': return 'text-green-300'
    case 'error': return 'text-red-300'
    case 'finish': return 'text-yellow-200'
    default: return 'text-gray-300'
  }
}

function getStatusColor(status: string): string {
  switch (status) {
    case 'running': return 'text-blue-700'
    case 'success': return 'text-green-700'
    case 'error': return 'text-red-700'
    default: return 'text-slate-500'
  }
}

function getStatusIcon(status: string): string {
  switch (status) {
    case 'running': return '🔄'
    case 'success': return '✓'
    case 'error': return '✗'
    default: return '⏳'
  }
}

// Actions
async function startGeneration() {
  store.startGeneration(props.databaseId)
}

function handleMinimize() {
  store.minimize()
  showModal.value = false
  emit('minimize')
  message.info('Task running in background.')
}

function handleCancel() {
  if (isForestRunning.value && store.chunkProgress.completedChunks > 0) {
    // Forest mode: completed chunks are already persisted — just stop remaining
    message.success(`Stopped. ${store.chunkProgress.completedChunks} completed clusters already saved.`)
  }
  store.cancelGeneration()
  showModal.value = false
}

function handleClose() {
  store.closeConsole()
  store.clearForestPreview()
  showModal.value = false
  if (store.isComplete) {
    emit('complete')
  }
}

// ----- Forest Preview (before generation) -----

// Fetch forest preview when console opens for large schemas
watch(() => props.show, (visible) => {
  if (visible && !store.isRunning && !store.isComplete && props.tableCount && props.tableCount > store.FOREST_THRESHOLD) {
    store.fetchForestPreview(props.databaseId)
  }
}, { immediate: true })

interface PreviewTreemapRect {
  cluster: ForestClusterPreview
  x: number
  y: number
  w: number
  h: number
}

function squarifyPreview(items: ForestClusterPreview[], containerW: number, containerH: number): PreviewTreemapRect[] {
  if (items.length === 0 || containerW <= 0 || containerH <= 0) return []
  const totalArea = containerW * containerH
  const totalValue = items.reduce((s, c) => s + Math.max(c.table_count, 1), 0)
  if (totalValue === 0) return []
  const areas: number[] = items.map(c => (Math.max(c.table_count, 1) / totalValue) * totalArea)
  const raw: { idx: number; x: number; y: number; w: number; h: number }[] = []
  function getArea(idx: number): number { return areas[idx] ?? 0 }
  function worst(row: number[], side: number): number {
    if (row.length === 0 || side <= 0) return Infinity
    let sum = 0, lo = Infinity, hi = -Infinity
    for (const idx of row) { const a = getArea(idx); sum += a; if (a < lo) lo = a; if (a > hi) hi = a }
    if (sum === 0) return Infinity
    const s2 = side * side, r2 = sum * sum
    return Math.max((s2 * hi) / r2, r2 / (s2 * lo))
  }
  const order: number[] = Array.from({ length: items.length }, (_, i) => i).sort((a, b) => getArea(b) - getArea(a))
  function recurse(rem: number[], x: number, y: number, w: number, h: number) {
    if (rem.length === 0) return
    if (rem.length === 1) { raw.push({ idx: rem[0]!, x, y, w, h }); return }
    const isWide = w >= h
    const side = isWide ? h : w
    const first = rem[0]!
    const row: number[] = [first]
    let rowArea = getArea(first)
    let best = worst(row, side)
    let i = 1
    for (; i < rem.length; i++) {
      const c = rem[i]!
      const nr = [...row, c]
      const nw = worst(nr, side)
      if (nw <= best) { row.push(c); rowArea += getArea(c); best = nw } else { break }
    }
    if (isWide) {
      const colW = rowArea / h; let cy = y
      for (const idx of row) { const itemA = getArea(idx); const itemH = itemA / colW; raw.push({ idx, x, y: cy, w: colW, h: itemH }); cy += itemH }
      recurse(rem.slice(i), x + colW, y, Math.max(w - colW, 0), h)
    } else {
      const rowH = rowArea / w; let cx = x
      for (const idx of row) { const itemA = getArea(idx); const itemW = itemA / rowH; raw.push({ idx, x: cx, y, w: itemW, h: rowH }); cx += itemW }
      recurse(rem.slice(i), x, y + rowH, w, Math.max(h - rowH, 0))
    }
  }
  recurse(order, 0, 0, containerW, containerH)
  const half = GAP / 2
  return raw.map(r => ({ cluster: items[r.idx]!, x: r.x + half, y: r.y + half, w: Math.max(r.w - GAP, 1), h: Math.max(r.h - GAP, 1) }))
}

const PREVIEW_TREEMAP_W = 760
const PREVIEW_TREEMAP_H = 180

const previewTreemapRects = computed(() => {
  const fp = store.forestPreview
  if (!fp || !fp.clusters || fp.clusters.length === 0) return []
  return squarifyPreview(fp.clusters, PREVIEW_TREEMAP_W, PREVIEW_TREEMAP_H)
})

function previewCellBg(cluster: ForestClusterPreview): string {
  return cluster.will_skip ? '#bae6fd' : '#fde68a' // sky-200 for skip, amber-200 for need
}
function previewCellBorder(cluster: ForestClusterPreview): string {
  return cluster.will_skip ? '#38bdf8' : '#f59e0b' // sky-400 for skip, amber-500 for need
}
function previewCellText(cluster: ForestClusterPreview): string {
  return cluster.will_skip ? '#0c4a6e' : '#78350f' // sky-900, amber-900
}
function previewCellLabel(rect: PreviewTreemapRect): string {
  const { cluster, w, h } = rect
  const idx = `#${cluster.index + 1}`
  const pct = Math.round(cluster.coverage_ratio * 100)
  if (w >= 80 && h >= 36) return `${idx} · ${cluster.table_count}t\n${pct}%`
  if (w >= 50 && h >= 24) return `${idx} · ${cluster.table_count}t`
  if (w >= 22 && h >= 16) return idx
  return ''
}

// Preview tooltip
const hoveredPreviewCluster = ref<ForestClusterPreview | null>(null)
const previewTooltipPos = ref({ x: 0, y: 0 })
function onPreviewCellEnter(e: MouseEvent, cluster: ForestClusterPreview) {
  hoveredPreviewCluster.value = cluster
  previewTooltipPos.value = { x: e.clientX, y: e.clientY }
}
function onPreviewCellMove(e: MouseEvent) {
  previewTooltipPos.value = { x: e.clientX, y: e.clientY }
}
function onPreviewCellLeave() {
  hoveredPreviewCluster.value = null
}
</script>

<template>
  <NModal
    v-model:show="showModal"
    preset="card"
    :closable="true"
    :mask-closable="!store.isRunning"
    :on-close="store.isRunning ? handleMinimize : undefined"
    style="width: 850px; max-width: 90vw;"
    class="generate-console-modal"
  >
    <template #header>
      <div class="flex items-center gap-2">
        <span class="text-slate-800 font-bold">Rich Context Generation</span>
        <span v-if="store.isRunning" class="px-2 py-0.5 text-xs font-semibold rounded-full bg-blue-100 text-blue-700 animate-pulse border border-blue-200">
          Running
        </span>
        <span v-else-if="store.isComplete" class="px-2 py-0.5 text-xs font-semibold rounded-full bg-green-100 text-green-700 border border-green-200">
          Complete
        </span>
      </div>
    </template>

    <div class="generate-console">
      <!-- Configuration (shown before start) -->
      <div v-if="!store.isRunning && !store.isComplete" class="config-section mb-6">
        <!-- Forest mode banner for large schemas -->
        <div v-if="props.tableCount && props.tableCount > store.FOREST_THRESHOLD" class="mb-4 p-4 rounded-lg bg-indigo-50 border border-indigo-200">
          <div class="flex items-center gap-2 mb-2">
            <span class="text-lg">🌲</span>
            <span class="font-bold text-indigo-900 text-sm">Forest-Based Chunked Onboarding</span>
          </div>
          <p class="text-xs text-indigo-800 leading-relaxed font-medium">
            {{ props.tableCount }} tables detected — exceeds the {{ store.FOREST_THRESHOLD }}-table threshold.
            The schema will be decomposed into FK-connected clusters, each onboarded independently.
          </p>
          <div class="mt-3 grid grid-cols-3 gap-3 text-center">
            <div class="bg-white rounded-md p-2 border border-indigo-100">
              <div class="text-lg font-bold text-indigo-900">{{ props.tableCount }}</div>
              <div class="text-[10px] font-semibold text-indigo-600">Total Tables</div>
            </div>
            <div class="bg-white rounded-md p-2 border border-indigo-100">
              <div class="text-lg font-bold text-indigo-900">Auto</div>
              <div class="text-[10px] font-semibold text-indigo-600">Iter / Chunk</div>
            </div>
            <div class="bg-white rounded-md p-2 border border-indigo-100">
              <div class="text-lg font-bold text-indigo-900">150</div>
              <div class="text-[10px] font-semibold text-indigo-600">Max / Chunk</div>
            </div>
          </div>
          <p class="mt-2 text-[11px] text-indigo-700 leading-relaxed">
            Each chunk's budget is computed as <code class="bg-indigo-100 px-1 rounded text-[10px]">3 × tables + 10</code>, capped at 150.
            For example: a 10-table cluster ≈ 24–60 iters, a 70-table cluster ≈ 132–150 iters.
          </p>
        </div>

        <!-- Forest Preview Treemap (loaded from backend) -->
        <div v-if="store.forestPreviewLoading" class="mb-4 p-4 rounded-lg bg-slate-50 border border-slate-200 text-center">
          <span class="text-sm text-slate-600 font-medium">🔍 Analyzing schema decomposition...</span>
        </div>
        <div v-else-if="store.forestPreviewError" class="mb-4 p-3 rounded-lg bg-red-50 border border-red-200">
          <span class="text-xs text-red-700 font-medium">⚠️ Failed to load preview: {{ store.forestPreviewError }}</span>
        </div>
        <div v-else-if="store.forestPreview" class="mb-4 p-4 rounded-lg bg-white border border-slate-200 shadow-sm">
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center gap-2">
              <span class="text-base">🗺️</span>
              <span class="font-bold text-sm text-slate-800">Chunk Preview</span>
            </div>
            <div class="flex items-center gap-3 text-xs font-bold">
              <span class="px-2 py-0.5 rounded-full bg-amber-100 text-amber-800 border border-amber-300">
                {{ store.forestPreview.clusters_need }} to generate
              </span>
              <span v-if="store.forestPreview.clusters_skip > 0" class="px-2 py-0.5 rounded-full bg-sky-100 text-sky-800 border border-sky-300">
                {{ store.forestPreview.clusters_skip }} will skip
              </span>
              <span class="text-slate-500">
                {{ store.forestPreview.clusters_total }} total
              </span>
            </div>
          </div>

          <!-- Preview Treemap -->
          <div class="treemap-wrap">
            <div class="treemap-container" :style="{ width: PREVIEW_TREEMAP_W + 'px', height: PREVIEW_TREEMAP_H + 'px' }">
              <div
                v-for="(rect, i) in previewTreemapRects"
                :key="i"
                class="treemap-cell"
                :class="{ 'treemap-cell--skipped': rect.cluster.will_skip }"
                :style="{
                  left: rect.x + 'px',
                  top: rect.y + 'px',
                  width: rect.w + 'px',
                  height: rect.h + 'px',
                  '--bg': previewCellBg(rect.cluster),
                  '--border': previewCellBorder(rect.cluster),
                  '--text': previewCellText(rect.cluster),
                }"
                @mouseenter="(e) => onPreviewCellEnter(e, rect.cluster)"
                @mousemove="onPreviewCellMove"
                @mouseleave="onPreviewCellLeave"
              >
                <span v-if="previewCellLabel(rect)" class="treemap-label">{{ previewCellLabel(rect) }}</span>
              </div>
            </div>
          </div>

          <!-- Preview Tooltip -->
          <Teleport to="body">
            <Transition name="tip">
              <div
                v-if="hoveredPreviewCluster"
                class="treemap-tooltip"
                :style="{ left: previewTooltipPos.x + 14 + 'px', top: previewTooltipPos.y + 14 + 'px' }"
              >
                <div class="flex items-center gap-2 mb-1">
                  <span class="font-bold text-xs text-slate-800">Cluster #{{ hoveredPreviewCluster.index + 1 }}</span>
                  <span class="text-[10px] px-1.5 py-0.5 rounded-full font-bold"
                    :class="hoveredPreviewCluster.will_skip ? 'bg-sky-100 text-sky-700' : 'bg-amber-100 text-amber-700'"
                  >{{ hoveredPreviewCluster.will_skip ? 'will skip' : 'needs generation' }}</span>
                </div>
                <div class="text-xs font-semibold text-slate-600">
                  {{ hoveredPreviewCluster.table_count }} tables · {{ hoveredPreviewCluster.relation_count }} FK
                </div>
                <div class="text-[10px] text-slate-500 mt-0.5">
                  Coverage: {{ Math.round(hoveredPreviewCluster.coverage_ratio * 100) }}% ({{ hoveredPreviewCluster.tables_with_context }}/{{ hoveredPreviewCluster.table_count }} tables)
                  <span v-if="hoveredPreviewCluster.coverage_ratio >= 0.9" class="text-sky-600 font-bold">✓ ≥90%</span>
                  <span v-else class="text-amber-600 font-bold">&lt;90%</span>
                </div>
                <div class="text-[10px] text-slate-500">
                  Columns: {{ hoveredPreviewCluster.columns_with_context }}/{{ hoveredPreviewCluster.columns_total }}
                </div>
                <div class="text-[10px] text-slate-500">
                  Budget: {{ hoveredPreviewCluster.min_iter }}–{{ hoveredPreviewCluster.max_iter }} iterations
                </div>
                <div v-if="hoveredPreviewCluster.tables.length > 0" class="text-[10px] font-medium text-slate-500 mt-1 max-w-56 leading-tight">
                  {{ hoveredPreviewCluster.tables.length <= 6 ? hoveredPreviewCluster.tables.join(', ') : hoveredPreviewCluster.tables.slice(0, 5).join(', ') + ` … +${hoveredPreviewCluster.tables.length - 5}` }}
                </div>
              </div>
            </Transition>
          </Teleport>

          <!-- Preview Legend -->
          <div class="flex items-center justify-between mt-3">
            <div class="flex gap-3 text-xs font-bold text-slate-700">
              <span class="flex items-center gap-1.5"><span class="legend-dot" style="background: #fde68a; border-color: #f59e0b;" /> Needs Generation</span>
              <span class="flex items-center gap-1.5"><span class="legend-dot" style="background: #bae6fd; border-color: #38bdf8;" /> Will Skip (has context)</span>
            </div>
            <div class="flex gap-3 text-xs font-bold text-slate-500">
              <span>Largest {{ store.forestPreview.largest_cluster }}t</span>
              <span>Median {{ store.forestPreview.median_cluster }}t</span>
              <span v-if="store.forestPreview.isolated_tables > 0">Isolated {{ store.forestPreview.isolated_tables }}</span>
            </div>
          </div>
        </div>

        <!-- Single-agent mode: show manual iteration config -->
        <template v-if="!props.tableCount || props.tableCount <= store.FOREST_THRESHOLD">
          <div class="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label class="text-sm font-semibold text-slate-700 mb-2 block">Min Iterations</label>
              <NInputNumber v-model:value="store.config.minIterations" :min="1" :max="store.config.maxIterations" size="small" />
            </div>
            <div>
              <label class="text-sm font-semibold text-slate-700 mb-2 block">Max Iterations</label>
              <NInputNumber v-model:value="store.config.maxIterations" :min="store.config.minIterations" :max="300" size="small" />
            </div>
          </div>
          <p v-if="props.tableCount && props.tableCount > 0" class="text-xs font-semibold text-blue-700 mb-3">
            💡 Recommended for {{ props.tableCount }} tables: {{ store.computeRecommendedIterations(props.tableCount).min }}–{{ store.computeRecommendedIterations(props.tableCount).max }} iterations
          </p>
        </template>

        <div class="flex items-center gap-4 mb-4">
          <div class="flex items-center gap-2">
            <NSwitch v-model:value="store.config.force" size="small" />
            <span class="text-sm font-semibold text-slate-700">Force Regenerate</span>
          </div>
        </div>
        <p class="text-xs font-medium text-slate-600 mb-4">
          The agent explores tables, discovers data patterns, and saves descriptions, sample values, synonyms, and business terms.
        </p>
        <NButton type="primary" size="large" class="w-full" @click="startGeneration">
          <template #icon><div class="i-lucide-play" /></template>
          Start Generation
        </NButton>
      </div>

      <!-- Running/Complete view -->
      <div v-else>
        <!-- Forest Treemap Visualization (shown only in forest mode) -->
        <div v-if="store.chunkProgress.isForestMode" class="chunk-progress-section mb-4 p-3 rounded-lg bg-slate-100 border border-slate-300 shadow-inner">
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center gap-2">
              <span class="text-base">🌲</span>
              <span class="font-bold text-sm" style="color: #1a1a1a;">Forest Chunked Mode</span>
            </div>
            <div class="flex items-center gap-3 text-xs">
              <span class="font-bold" style="color: #1a1a1a;">
                {{ store.chunkProgress.completedChunks }}/{{ store.chunkProgress.clustersTotal }} clusters
              </span>
              <span v-if="store.chunkProgress.erroredChunks > 0" class="text-red-600 font-bold">({{ store.chunkProgress.erroredChunks }} err)</span>
            </div>
          </div>

          <!-- Treemap -->
          <div class="treemap-wrap">
            <div class="treemap-container" :style="{ width: TREEMAP_W + 'px', height: TREEMAP_H + 'px' }">
              <div
                v-for="(rect, i) in treemapRects"
                :key="i"
                class="treemap-cell"
                :class="{
                  'treemap-cell--running': rect.cluster.status === 'running',
                  'treemap-cell--done':    rect.cluster.status === 'success',
                  'treemap-cell--skipped': rect.cluster.status === 'skipped',
                  'treemap-cell--error':   rect.cluster.status === 'error',
                }"
                :style="{
                  left: rect.x + 'px',
                  top: rect.y + 'px',
                  width: rect.w + 'px',
                  height: rect.h + 'px',
                  '--bg': cellBg(rect.cluster.status),
                  '--border': cellBorder(rect.cluster.status),
                  '--text': cellText(rect.cluster.status),
                }"
                @mouseenter="(e) => onCellEnter(e, rect.cluster)"
                @mousemove="onCellMove"
                @mouseleave="onCellLeave"
              >
                <span v-if="cellLabel(rect)" class="treemap-label">{{ cellLabel(rect) }}</span>
              </div>
            </div>
          </div>

          <!-- Tooltip (follows cursor) -->
          <Teleport to="body">
            <Transition name="tip">
              <div
                v-if="hoveredCluster"
                class="treemap-tooltip"
                :style="{ left: tooltipPos.x + 14 + 'px', top: tooltipPos.y + 14 + 'px' }"
              >
                <div class="flex items-center gap-2 mb-1">
                  <span class="font-bold text-xs text-slate-800">Cluster #{{ hoveredCluster.index + 1 }}</span>
                  <span class="text-[10px] px-1.5 py-0.5 rounded-full font-bold"
                    :class="{
                      'bg-green-100 text-green-700': hoveredCluster.status === 'success',
                      'bg-sky-100 text-sky-700': hoveredCluster.status === 'skipped',
                      'bg-amber-100 text-amber-700': hoveredCluster.status === 'running',
                      'bg-red-100 text-red-700': hoveredCluster.status === 'error',
                      'bg-slate-200 text-slate-700': hoveredCluster.status === 'pending',
                    }"
                  >{{ hoveredCluster.status }}</span>
                </div>
                <div class="text-xs font-semibold text-slate-600">
                  {{ hoveredCluster.tableCount }} tables · {{ hoveredCluster.relationCount }} FK
                </div>
                <div v-if="hoveredCluster.coverageRatio > 0" class="text-[10px] text-slate-500 mt-0.5">
                  Coverage: {{ Math.round(hoveredCluster.coverageRatio * 100) }}%
                  <span v-if="hoveredCluster.coverageRatio >= 0.9" class="text-sky-600 font-bold">✓ ≥90%</span>
                </div>
                <div v-if="hoveredCluster.tables.length > 0" class="text-[10px] font-medium text-slate-500 mt-1 max-w-56 leading-tight">
                  {{ hoveredCluster.tables.length <= 6 ? hoveredCluster.tables.join(', ') : hoveredCluster.tables.slice(0, 5).join(', ') + ` … +${hoveredCluster.tables.length - 5}` }}
                </div>
              </div>
            </Transition>
          </Teleport>

          <!-- Legend + Stats row -->
          <div class="flex items-center justify-between mt-3">
            <div class="flex gap-3 text-xs font-bold" style="color: #1a1a1a;">
              <span class="flex items-center gap-1.5"><span class="legend-dot" :style="{ background: cellBg('success'), borderColor: cellBorder('success') }" /> Done</span>
              <span class="flex items-center gap-1.5"><span class="legend-dot" :style="{ background: cellBg('skipped'), borderColor: cellBorder('skipped') }" /> Skipped</span>
              <span class="flex items-center gap-1.5"><span class="legend-dot legend-dot--pulse" :style="{ background: cellBg('running'), borderColor: cellBorder('running') }" /> Active</span>
              <span class="flex items-center gap-1.5"><span class="legend-dot" :style="{ background: cellBg('pending'), borderColor: cellBorder('pending') }" /> Pending</span>
              <span class="flex items-center gap-1.5"><span class="legend-dot" :style="{ background: cellBg('error'), borderColor: cellBorder('error') }" /> Error</span>
            </div>
            <div class="flex gap-3 text-xs font-bold" style="color: #1a1a1a;">
              <span>Largest {{ store.chunkProgress.largestCluster }}t</span>
              <span>Median {{ store.chunkProgress.medianCluster }}t</span>
              <span v-if="store.chunkProgress.isolatedTables > 0">Isolated {{ store.chunkProgress.isolatedTables }}</span>
            </div>
          </div>

          <!-- Current chunk detail card -->
          <div v-if="store.chunkProgress.currentChunk >= 0 && store.chunkProgress.completedChunks < store.chunkProgress.clustersTotal"
            class="mt-3 px-3 py-2.5 rounded-md bg-amber-50 border border-amber-300 shadow-sm">
            <div class="text-xs text-amber-900 font-bold">
              ▶ Chunk {{ store.chunkProgress.currentChunk + 1 }}/{{ store.chunkProgress.clustersTotal }}
              — {{ store.chunkProgress.currentChunkTableCount }} tables, {{ store.chunkProgress.currentChunkRelationCount }} FK
              <span v-if="store.agentState.iteration > 0" class="ml-1 text-amber-700 font-medium">· iter {{ store.agentState.iteration }}</span>
            </div>
            <div v-if="store.chunkProgress.currentChunkTables.length > 0" class="text-[11px] text-slate-800 mt-1 truncate">
              {{ store.chunkProgress.currentChunkTables.length <= 8 ? store.chunkProgress.currentChunkTables.join(', ') : store.chunkProgress.currentChunkTables.slice(0, 6).join(', ') + ` … +${store.chunkProgress.currentChunkTables.length - 6}` }}
            </div>
          </div>
        </div>

        <!-- Agent Status -->
        <div class="agent-status-section mb-4">
          <h4 class="text-sm font-semibold text-slate-700 mb-3">Agent Status</h4>
          <div class="grid grid-cols-2 gap-3">
            <!-- RC Gen Agent -->
            <div class="agent-card p-3 rounded-lg bg-blue-50 border border-blue-200">
              <div class="flex items-center justify-between mb-2">
                <div class="flex items-center gap-2">
                  <span class="text-lg">🤖</span>
                  <span class="font-bold text-blue-900 text-sm">RC Generator</span>
                </div>
                <span :class="['text-xs font-semibold', getStatusColor(store.agentState.status)]">
                  {{ getStatusIcon(store.agentState.status) }} {{ store.agentState.status }}
                </span>
              </div>
              <NProgress
                :percentage="store.agentState.progress"
                :show-indicator="false"
                :color="store.agentState.status === 'success' ? '#16a34a' : store.agentState.status === 'error' ? '#dc2626' : '#2563eb'"
                :height="6"
              />
              <div class="text-xs text-slate-700 font-medium mt-1">
                <template v-if="store.chunkProgress.isForestMode && store.chunkProgress.currentChunk >= 0">
                  <span>Chunk {{ store.chunkProgress.currentChunk + 1 }}/{{ store.chunkProgress.clustersTotal }}</span>
                  <span v-if="store.agentState.iteration > 0" class="ml-2">· Iter {{ store.agentState.iteration }}</span>
                </template>
                <template v-else>
                  <span v-if="store.agentState.iteration > 0">Iteration {{ store.agentState.iteration }}</span>
                  <span v-if="store.agentState.phase" class="ml-2">· {{ store.agentState.phase }}</span>
                </template>
              </div>
            </div>

            <!-- Embedding Agent -->
            <div class="agent-card p-3 rounded-lg bg-purple-50 border border-purple-200">
              <div class="flex items-center justify-between mb-2">
                <div class="flex items-center gap-2">
                  <span class="text-lg">🧬</span>
                  <span class="font-bold text-purple-900 text-sm">Embeddings</span>
                </div>
                <span :class="['text-xs font-semibold', getStatusColor(store.embeddingState.status)]">
                  {{ getStatusIcon(store.embeddingState.status) }} {{ store.embeddingState.status }}
                </span>
              </div>
              <NProgress
                :percentage="store.embeddingState.progress"
                :show-indicator="false"
                :color="store.embeddingState.status === 'success' ? '#16a34a' : store.embeddingState.status === 'error' ? '#dc2626' : '#9333ea'"
                :height="6"
              />
              <div class="text-xs text-slate-700 font-medium mt-1">
                {{ store.embeddingState.message }}
              </div>
              <div v-if="store.storageStats.embeddingsStreamed > 0" class="text-[10px] text-purple-600 font-semibold mt-0.5">
                🧬 {{ store.storageStats.embeddingsStreamed }} vectors written
              </div>
            </div>
          </div>
        </div>

        <!-- Storage Stats -->
        <div class="storage-section mb-4 p-3 rounded-lg bg-emerald-50 border border-emerald-200">
          <h4 class="text-sm font-bold text-emerald-900 mb-2 flex items-center gap-2">
            <span>💾</span> Storage Activity
            <span class="text-xs text-slate-700 font-semibold ml-auto">{{ store.totalContextWrites }} writes</span>
          </h4>
          <div class="grid grid-cols-2 gap-4 mb-2">
            <div>
              <div class="text-xs font-semibold text-slate-700 mb-1">Table Descriptions</div>
              <div class="dual-progress-bar" :style="{ '--total': store.storageStats.tablesTotal || 1 }">
                <div class="dual-progress-track">
                  <div class="dual-progress-existed" :style="{ width: store.storageStats.tablesTotal ? (store.storageStats.tablesExisting / store.storageStats.tablesTotal * 100) + '%' : '0%' }" />
                  <div class="dual-progress-new" :style="{ width: store.storageStats.tablesTotal ? (store.storageStats.tablesUpdated / store.storageStats.tablesTotal * 100) + '%' : '0%' }" />
                </div>
                <span class="text-xs font-bold text-slate-800 ml-2 whitespace-nowrap">
                  {{ store.storageStats.tablesUpdated }}/{{ store.storageStats.tablesTotal }}
                  <span v-if="store.storageStats.tablesExisting > 0" class="text-slate-500 font-normal ml-0.5">({{ store.storageStats.tablesExisting }} existed)</span>
                </span>
              </div>
            </div>
            <div>
              <div class="text-xs font-semibold text-slate-700 mb-1">Column Descriptions</div>
              <div class="dual-progress-bar" :style="{ '--total': store.storageStats.columnsTotal || 1 }">
                <div class="dual-progress-track">
                  <div class="dual-progress-existed" :style="{ width: store.storageStats.columnsTotal ? (store.storageStats.columnsExisting / store.storageStats.columnsTotal * 100) + '%' : '0%' }" />
                  <div class="dual-progress-new" :style="{ width: store.storageStats.columnsTotal ? (store.storageStats.columnsUpdated / store.storageStats.columnsTotal * 100) + '%' : '0%' }" />
                </div>
                <span class="text-xs font-bold text-slate-800 ml-2 whitespace-nowrap">
                  {{ store.storageStats.columnsUpdated }}/{{ store.storageStats.columnsTotal }}
                  <span v-if="store.storageStats.columnsExisting > 0" class="text-slate-500 font-normal ml-0.5">({{ store.storageStats.columnsExisting }} existed)</span>
                </span>
              </div>
            </div>
          </div>
          <div class="flex gap-4 text-xs font-semibold text-slate-700">
            <span v-if="store.storageStats.sampleValuesAdded > 0">📋 Sample Values: {{ store.storageStats.sampleValuesAdded }}</span>
            <span v-if="store.storageStats.synonymsAdded > 0">🔗 Synonyms: {{ store.storageStats.synonymsAdded }}</span>
            <span v-if="store.storageStats.termsAdded > 0">📖 Terms: {{ store.storageStats.termsAdded }}</span>
          </div>
        </div>

        <!-- Console Output -->
        <div class="console-section">
          <h4 class="text-sm font-semibold text-slate-700 mb-2 flex items-center gap-2">
            <span>📋</span> Console Output
            <!-- Legend -->
            <span class="ml-auto flex items-center gap-3 text-xs font-semibold text-slate-600">
              <span class="flex items-center gap-1"><span>💭</span><span class="text-slate-500">Thought</span></span>
              <span class="flex items-center gap-1"><span>🔧</span><span class="text-blue-700">Action</span></span>
              <span class="flex items-center gap-1"><span>📊</span><span class="text-cyan-700">Result</span></span>
              <span class="flex items-center gap-1"><span>💾</span><span class="text-emerald-700">Saved</span></span>
            </span>
          </h4>
          <div class="console-log-area h-80 overflow-y-auto bg-slate-900 rounded-lg p-3 font-mono text-xs leading-relaxed">
            <div v-for="log in store.logs" :key="log.id" class="log-entry py-0.5">
              <span class="text-slate-500">[{{ log.timestamp }}]</span>
              <span class="ml-1">{{ getPhaseIcon(log.phase) }}</span>
              <span :class="['ml-1', getPhaseColor(log.phase)]">{{ log.message }}</span>
              <div v-if="log.detail" class="ml-8 text-slate-400 break-all">{{ log.detail }}</div>
            </div>
            <div v-if="store.isRunning" class="cursor-blink inline-block w-2 h-4 bg-cyan-400 ml-1" />
          </div>
        </div>

        <!-- Footer Stats -->
        <div class="footer-stats mt-4 flex items-center justify-between text-sm font-semibold text-slate-700">
          <span>⏱️ {{ store.formattedElapsed }}</span>
          <span>📊 Tables: {{ store.storageStats.tablesUpdated }}/{{ store.storageStats.tablesTotal }}</span>
          <span>📝 Columns: {{ store.storageStats.columnsUpdated }}/{{ store.storageStats.columnsTotal }}</span>
          <span>💾 Writes: {{ store.totalContextWrites }}</span>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-between">
        <div>
          <NButton v-if="store.isRunning" quaternary size="small" @click="handleMinimize">
            <template #icon><span class="i-lucide-minimize-2" /></template>
            Minimize to Background
          </NButton>
        </div>
        <div class="flex gap-2">
          <template v-if="store.isRunning">
            <NButton v-if="isForestRunning" type="warning" size="small" @click="handleCancel">
              Stop Remaining
              <span v-if="store.chunkProgress.completedChunks > 0" class="ml-1 text-xs opacity-80">({{ store.chunkProgress.completedChunks }} saved)</span>
            </NButton>
            <NButton v-else type="error" size="small" @click="handleCancel">Cancel</NButton>
          </template>
          <NButton v-else size="small" @click="handleClose">Close</NButton>
        </div>
      </div>
    </template>
  </NModal>
</template>

<style scoped>
.generate-console-modal :deep(.n-card) {
  /* Using bright theme colors */
  background: #ffffff;
  border: 1px solid #e2e8f0;
}

.console-log-area::-webkit-scrollbar { width: 6px; }
.console-log-area::-webkit-scrollbar-track { background: rgba(255,255,255,0.05); border-radius: 3px; }
.console-log-area::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.2); border-radius: 3px; }

.cursor-blink { animation: blink 1s infinite; }
@keyframes blink { 0%,50%{opacity:1} 51%,100%{opacity:0} }

/* ---------- Treemap ---------- */
.treemap-wrap {
  border-radius: 8px;
  display: flex;
  justify-content: center;
  padding: 4px 0;
}
.treemap-container {
  position: relative;
  border-radius: 6px;
  background: transparent;
}

.treemap-cell {
  position: absolute;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--bg);
  border: 1px solid var(--border);
  transition: background-color 0.5s, border-color 0.5s, box-shadow 0.5s;
  cursor: default;
  overflow: hidden;
}
.treemap-cell:hover {
  filter: brightness(1.3);
  z-index: 2;
}

/* Running: amber glow pulse */
.treemap-cell--running {
  animation: cell-glow 2s ease-in-out infinite;
}
@keyframes cell-glow {
  0%,100% { box-shadow: 0 0 4px rgba(245,158,11,0.5); }
  50%     { box-shadow: 0 0 12px rgba(245,158,11,0.9); }
}

/* Done: subtle inner glow */
.treemap-cell--done {
  box-shadow: inset 0 0 6px rgba(255,255,255,0.4);
}

/* Skipped: diagonal stripe pattern to visually distinguish from "done" */
.treemap-cell--skipped {
  box-shadow: inset 0 0 6px rgba(255,255,255,0.4);
  background-image: repeating-linear-gradient(
    135deg,
    transparent,
    transparent 3px,
    rgba(255,255,255,0.35) 3px,
    rgba(255,255,255,0.35) 5px
  );
}

/* Error: red inner glow */
.treemap-cell--error {
  box-shadow: inset 0 0 6px rgba(255,255,255,0.4);
}

.treemap-label {
  font-size: 11px;
  font-weight: 700;
  line-height: 1.2;
  color: var(--text);
  text-shadow: 0 1px 1px rgba(255,255,255,0.4);
  pointer-events: none;
  user-select: none;
  white-space: pre-line;
  text-align: center;
}

/* Legend dots */
.legend-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 3px;
  border: 1px solid;
}
.legend-dot--pulse {
  animation: cell-glow 2s ease-in-out infinite;
}

/* Tooltip */
.treemap-tooltip {
  position: fixed;
  z-index: 9999;
  background: rgba(255,255,255,0.96);
  border: 1px solid rgba(203,213,225,1); /* slate-300 */
  border-radius: 8px;
  padding: 8px 12px;
  pointer-events: none;
  box-shadow: 0 8px 24px rgba(0,0,0,0.15);
  backdrop-filter: blur(12px);
  max-width: 260px;
}

/* Tooltip enter/leave transition */
.tip-enter-active { transition: opacity 0.15s ease; }
.tip-leave-active { transition: opacity 0.1s ease; }
.tip-enter-from, .tip-leave-to { opacity: 0; }

/* ---------- Dual-color progress bar ---------- */
.dual-progress-bar {
  display: flex;
  align-items: center;
}
.dual-progress-track {
  position: relative;
  flex: 1;
  height: 8px;
  border-radius: 4px;
  background: #e2e8f0; /* slate-200 */
  overflow: hidden;
}
/* Existed layer: light teal, sits behind the "new" layer */
.dual-progress-existed {
  position: absolute;
  left: 0;
  top: 0;
  height: 100%;
  background: #99f6e4; /* teal-200 */
  border-radius: 4px;
  transition: width 0.4s ease;
}
/* New layer: solid green, overlaps from 0 to updatedPct */
.dual-progress-new {
  position: absolute;
  left: 0;
  top: 0;
  height: 100%;
  background: #059669; /* emerald-600 */
  border-radius: 4px;
  transition: width 0.4s ease;
}
</style>
