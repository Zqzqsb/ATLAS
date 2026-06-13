<script setup lang="ts">
/**
 * Accelerated demo of the Coordinator's forest-decomposed chunks being
 * dispatched and processed by Workers. Self-contained: synthetic clusters +
 * squarified treemap layout (Bruls–Huizing–van Wijk) + looping fill animation.
 * Mirrors the real Chunk Preview in ContextManager/GenerateContextConsole.vue.
 */
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'

type Status = 'pending' | 'running' | 'done' | 'skip'
interface Cluster { index: number; t: number; status: Status }

// Synthetic distribution resembling a real large DB (~26 clusters).
const SIZES = [70, 35, 25, 25, 25, 25, 25, 25, 20, 20, 20, 20, 20, 20, 20, 20, 20, 19, 14, 10, 10, 10, 10, 8, 8, 8]
// A few clusters already have context → start "skipped" (mixed demo).
const SKIP = new Set([3, 7, 11, 16, 20, 24])

const clusters = reactive<Cluster[]>(
  SIZES.map((t, i) => ({ index: i, t, status: SKIP.has(i) ? 'skip' : 'pending' })),
)

const W = 420
const H = 300
const GAP = 3

interface Rect { c: Cluster; x: number; y: number; w: number; h: number }

function squarify(items: Cluster[], cw: number, ch: number): Rect[] {
  if (!items.length || cw <= 0 || ch <= 0) return []
  const totalArea = cw * ch
  const totalValue = items.reduce((s, c) => s + Math.max(c.t, 1), 0)
  const areas = items.map((c) => (Math.max(c.t, 1) / totalValue) * totalArea)
  const raw: { idx: number; x: number; y: number; w: number; h: number }[] = []
  const getArea = (i: number) => areas[i] ?? 0

  function worst(row: number[], side: number): number {
    if (!row.length || side <= 0) return Infinity
    let sum = 0, lo = Infinity, hi = -Infinity
    for (const i of row) { const a = getArea(i); sum += a; if (a < lo) lo = a; if (a > hi) hi = a }
    if (sum === 0) return Infinity
    const s2 = side * side, r2 = sum * sum
    return Math.max((s2 * hi) / r2, r2 / (s2 * lo))
  }

  const order = Array.from({ length: items.length }, (_, i) => i).sort((a, b) => getArea(b) - getArea(a))

  function recurse(rem: number[], x: number, y: number, w: number, h: number) {
    if (!rem.length) return
    if (rem.length === 1) { raw.push({ idx: rem[0]!, x, y, w, h }); return }
    const isWide = w >= h
    const side = isWide ? h : w
    const row: number[] = [rem[0]!]
    let rowArea = getArea(rem[0]!)
    let best = worst(row, side)
    let i = 1
    for (; i < rem.length; i++) {
      const nr = [...row, rem[i]!]
      const nw = worst(nr, side)
      if (nw <= best) { row.push(rem[i]!); rowArea += getArea(rem[i]!); best = nw } else break
    }
    if (isWide) {
      const colW = rowArea / h
      let cy = y
      for (const idx of row) { const ih = getArea(idx) / colW; raw.push({ idx, x, y: cy, w: colW, h: ih }); cy += ih }
      recurse(rem.slice(i), x + colW, y, Math.max(w - colW, 0), h)
    } else {
      const rowH = rowArea / w
      let cx = x
      for (const idx of row) { const iw = getArea(idx) / rowH; raw.push({ idx, x: cx, y, w: iw, h: rowH }); cx += iw }
      recurse(rem.slice(i), x, y + rowH, w, Math.max(h - rowH, 0))
    }
  }

  recurse(order, 0, 0, cw, ch)
  const half = GAP / 2
  return raw.map((r) => ({ c: items[r.idx]!, x: r.x + half, y: r.y + half, w: Math.max(r.w - GAP, 1), h: Math.max(r.h - GAP, 1) }))
}

const rects = computed(() => squarify(clusters, W, H))

function cellStyle(c: Cluster) {
  const map: Record<Status, [string, string, string]> = {
    done:    ['#86efac', '#22c55e', '#14532d'],
    skip:    ['#bae6fd', '#38bdf8', '#0c4a6e'],
    running: ['#fcd34d', '#f59e0b', '#78350f'],
    pending: ['#e2e8f0', '#cbd5e1', '#475569'],
  }
  const [bg, border, color] = map[c.status]
  return { background: bg, borderColor: border, color }
}
function label(r: Rect): string {
  const id = `#${r.c.index + 1}`
  if (r.w >= 50 && r.h >= 26) return `${id} · ${r.c.t}t`
  if (r.w >= 22 && r.h >= 15) return id
  return ''
}

// ─── Stats ───
const total = clusters.length
const needs = clusters.filter((c) => c.status !== 'skip').length
const skip = total - needs
const largest = Math.max(...SIZES)
const median = [...SIZES].sort((a, b) => a - b)[Math.floor(SIZES.length / 2)]
const isolated = SIZES.filter((t) => t <= 8).length
const doneCount = computed(() => clusters.filter((c) => c.status === 'done').length)
const runningIdx = computed(() => clusters.find((c) => c.status === 'running')?.index ?? -1)
const allDone = computed(() => doneCount.value >= needs)

// ─── Accelerated looping fill ───
const STEP_MS = 230
const HOLD_MS = 1600
let timer: number | null = null
let order: number[] = []
let ptr = 0

function reset() {
  for (const c of clusters) if (c.status !== 'skip') c.status = 'pending'
  order = clusters.filter((c) => c.status === 'pending').map((c) => c.index)
  ptr = 0
}
function step() {
  // complete the previous running cell
  const prev = clusters.find((c) => c.status === 'running')
  if (prev) prev.status = 'done'
  if (ptr >= order.length) {
    if (timer) { clearInterval(timer); timer = null }
    window.setTimeout(() => { reset(); start() }, HOLD_MS)
    return
  }
  const cur = clusters[order[ptr]!]
  if (cur) cur.status = 'running'
  ptr++
}
function start() {
  if (timer) clearInterval(timer)
  timer = window.setInterval(step, STEP_MS)
}

onMounted(() => { reset(); start() })
onUnmounted(() => { if (timer) clearInterval(timer) })

const legend = [
  { label: 'Needs Generation', cls: 'bg-slate-200 border-slate-400' },
  { label: 'Processing', cls: 'bg-amber-300 border-amber-500' },
  { label: 'Done', cls: 'bg-green-300 border-green-500' },
  { label: 'Will Skip', cls: 'bg-sky-200 border-sky-400' },
]
</script>

<template>
  <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
    <!-- header -->
    <div class="flex items-center gap-2 px-3.5 py-2.5 border-b border-slate-100">
      <div class="i-lucide-layout-grid text-slate-500 text-sm" />
      <span class="text-sm font-bold text-slate-800">Chunk Preview</span>
      <span class="text-[11px] text-slate-400 ml-1">加速演示</span>
      <div class="ml-auto flex items-center gap-1.5">
        <span class="px-2 py-0.5 rounded-full text-[11px] font-semibold bg-amber-50 text-amber-700 border border-amber-200">{{ needs }} to generate</span>
        <span class="px-2 py-0.5 rounded-full text-[11px] font-semibold bg-sky-50 text-sky-700 border border-sky-200">{{ skip }} will skip</span>
      </div>
    </div>

    <!-- treemap -->
    <div class="p-3">
      <div class="relative mx-auto" :style="{ width: W + 'px', height: H + 'px', maxWidth: '100%' }">
        <div
          v-for="r in rects"
          :key="r.c.index"
          class="absolute rounded-md border flex items-center justify-center text-center overflow-hidden transition-all duration-300"
          :class="{ 'chunk-pulse': r.c.status === 'running' }"
          :style="{ left: r.x + 'px', top: r.y + 'px', width: r.w + 'px', height: r.h + 'px', ...cellStyle(r.c) }"
        >
          <span class="text-[11px] font-semibold leading-none px-0.5 truncate">{{ label(r) }}</span>
        </div>
      </div>
    </div>

    <!-- footer: legend + stats -->
    <div class="flex items-center flex-wrap gap-x-3 gap-y-1.5 px-3.5 py-2.5 border-t border-slate-100">
      <div v-for="l in legend" :key="l.label" class="flex items-center gap-1 text-[11px] text-slate-500">
        <span class="w-2.5 h-2.5 rounded-sm border" :class="l.cls" />{{ l.label }}
      </div>
      <div class="ml-auto flex items-center gap-2 text-[11px] text-slate-400 font-mono">
        <span v-if="!allDone && runningIdx >= 0" class="text-amber-600 font-semibold">▶ #{{ runningIdx + 1 }}</span>
        <span v-else-if="allDone" class="text-green-600 font-semibold">✓ {{ doneCount }}/{{ needs }}</span>
        <span>Largest {{ largest }}t</span>
        <span>Median {{ median }}t</span>
        <span v-if="isolated > 0">Isolated {{ isolated }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.chunk-pulse {
  animation: chunk-pulse 0.9s ease-in-out infinite;
}
@keyframes chunk-pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(245, 158, 11, 0.4); }
  50% { box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.15); }
}
</style>
