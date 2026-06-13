<script setup lang="ts">
/**
 * Accelerated demo of the self-maintenance heal loop. Self-contained looping
 * animation driven by one playhead (0..100):
 *   detect → Coordinator marks RC expired → Executor heals per row → re-embed.
 * Each row is a Rich Context entry; the three soft-flag chips (is_expired /
 * is_stale / is_deleted) flip as it moves through its lifecycle:
 *   fresh → expired(E) → healing → healed(S, just rewritten) → re-embedded(clean)
 *   dropped column → expired → deleted(D).
 * Mirrors agent_service.ProcessSignal (Coordinator→Executor) + the soft flags in
 * lakebase repository.go / vector.go, and the closing GenerateAndSaveEmbeddings.
 */
import { ref, computed, onMounted, onUnmounted } from 'vue'

type RowState = 'fresh' | 'expired' | 'healing' | 'healed' | 'deleted'

interface Row {
  col: string
  /** change tag shown on the row */
  tag: string
  /** terminal action: create/refresh → heal to green; delete → deleted */
  act: 'create' | 'refresh' | 'delete'
}

// 4 affected RC entries. users.tier is downstream-affected (no direct DDL chip).
const ROWS: Row[] = [
  { col: 'orders.vip_level', tag: '新增列', act: 'create' },
  { col: 'orders.amount', tag: '类型变更', act: 'refresh' },
  { col: 'users.tier', tag: '受影响', act: 'refresh' },
  { col: 'orders.legacy_flag', tag: '删除列', act: 'delete' },
]
const N = ROWS.length

// the DDL signal chips (directly detected changes)
const CHANGES = [
  { sign: '+', col: 'orders.vip_level', cls: 'border-emerald-200 bg-emerald-50 text-emerald-700' },
  { sign: '~', col: 'orders.amount', cls: 'border-amber-200 bg-amber-50 text-amber-700' },
  { sign: '−', col: 'orders.legacy_flag', cls: 'border-rose-200 bg-rose-50 text-rose-700' },
]

// ─── playhead ───
const progress = ref(0)

// phase boundaries (global progress)
const DETECT_END = 16
const markPoint = (i: number) => DETECT_END + (i / N) * 16 // 16 → 32
const healStart = (i: number) => 36 + (i / N) * 34 // 36 → 70
const healDone = (i: number) => healStart(i) + 8
const embedPoint = (i: number) => 80 + (i / N) * 14 // 80 → 94

function rowState(i: number): RowState {
  const p = progress.value
  if (p < markPoint(i)) return 'fresh'
  if (p < healStart(i)) return 'expired'
  if (p < healDone(i)) return 'healing'
  return ROWS[i]!.act === 'delete' ? 'deleted' : 'healed'
}
function embedded(i: number): boolean {
  return ROWS[i]!.act !== 'delete' && progress.value >= embedPoint(i)
}

/** the 3 soft flags for a row, derived from its lifecycle state */
function flags(i: number): { e: boolean; s: boolean; d: boolean } {
  const st = rowState(i)
  if (st === 'expired' || st === 'healing') return { e: true, s: false, d: false }
  if (st === 'healed') return { e: false, s: !embedded(i), d: false }
  if (st === 'deleted') return { e: false, s: false, d: true }
  return { e: false, s: false, d: false }
}

const rowCls: Record<RowState, string> = {
  fresh: 'border-gray-200 bg-white',
  expired: 'border-amber-300 bg-amber-50/70',
  healing: 'border-amber-400 bg-amber-50',
  healed: 'border-emerald-300 bg-emerald-50/70',
  deleted: 'border-gray-200 bg-gray-50 opacity-60',
}

// ─── overall phase / status ───
const phase = computed(() => {
  const p = progress.value
  if (p < DETECT_END) return { key: 'detect', label: 'Detecting…', agent: '', cls: 'bg-amber-50 text-amber-700 border-amber-200' }
  if (p < 34) return { key: 'mark', label: 'Marking expired…', agent: 'Coordinator', cls: 'bg-violet-50 text-violet-700 border-violet-200' }
  if (p < 78) return { key: 'heal', label: 'Healing…', agent: 'Executor', cls: 'bg-emerald-50 text-emerald-700 border-emerald-200' }
  if (p < 96) return { key: 'embed', label: 'Re-embedding…', agent: '', cls: 'bg-indigo-50 text-indigo-700 border-indigo-200' }
  return { key: 'done', label: 'Synced', agent: '', cls: 'bg-emerald-50 text-emerald-700 border-emerald-200' }
})

const PHASES = [
  { key: 'detect', label: '检测', icon: 'i-lucide-search' },
  { key: 'mark', label: '标失效', icon: 'i-lucide-flag' },
  { key: 'heal', label: '自愈', icon: 'i-lucide-wand-2' },
  { key: 'embed', label: '重嵌入', icon: 'i-lucide-radar' },
] as const
const phaseIdx = computed(() => PHASES.findIndex((p) => p.key === phase.value.key))

// ─── looping playhead ───
const STEP = 1.3
const TICK_MS = 30
const HOLD_MS = 1500
let timer: number | null = null
function tick() {
  progress.value += STEP
  if (progress.value >= 100) {
    progress.value = 100
    if (timer) { clearInterval(timer); timer = null }
    window.setTimeout(() => { progress.value = 0; start() }, HOLD_MS)
  }
}
function start() {
  if (timer) clearInterval(timer)
  timer = window.setInterval(tick, TICK_MS)
}
onMounted(start)
onUnmounted(() => { if (timer) clearInterval(timer) })
</script>

<template>
  <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
    <!-- header -->
    <div class="flex items-center gap-2 px-3.5 py-2.5 border-b border-slate-100">
      <div class="i-lucide-heart-pulse text-amber-500 text-sm" />
      <span class="text-sm font-bold text-slate-800">Rich Context 自愈</span>
      <span class="text-[11px] text-slate-400 ml-1">加速演示</span>
      <span class="ml-auto px-2 py-0.5 rounded-full text-[11px] font-semibold border transition-colors" :class="phase.cls">
        <span v-if="phase.agent" class="font-mono opacity-70 mr-1">{{ phase.agent }}</span>{{ phase.label }}
      </span>
    </div>

    <!-- phase track -->
    <div class="flex items-center gap-1 px-3.5 pt-2.5">
      <template v-for="(ph, i) in PHASES" :key="ph.key">
        <div
          class="flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold border transition-colors duration-300"
          :class="i <= phaseIdx ? 'border-slate-300 bg-slate-50 text-slate-700' : 'border-gray-100 bg-gray-50/50 text-gray-300'"
        >
          <div :class="[ph.icon, 'text-[11px]']" />{{ ph.label }}
        </div>
        <div v-if="i < PHASES.length - 1" class="i-lucide-chevron-right flex-shrink-0 text-xs" :class="i < phaseIdx ? 'text-slate-300' : 'text-gray-200'" />
      </template>
    </div>

    <!-- signal chips -->
    <div class="flex flex-wrap items-center gap-1.5 px-3.5 pt-2.5">
      <span class="text-[10px] text-slate-400 font-semibold">DDL</span>
      <code
        v-for="c in CHANGES"
        :key="c.col"
        class="px-1.5 py-0.5 rounded-md border text-[10px] font-mono transition-opacity duration-300"
        :class="[c.cls, progress >= DETECT_END * 0.4 ? 'opacity-100' : 'opacity-30']"
      ><b>{{ c.sign }}</b> {{ c.col }}</code>
    </div>

    <!-- RC lifecycle rows -->
    <div class="px-3.5 py-2.5 space-y-1.5">
      <div
        v-for="(r, i) in ROWS"
        :key="r.col"
        class="flex items-center gap-2 rounded-lg border px-2.5 py-1.5 transition-all duration-300"
        :class="[rowCls[rowState(i)], { 'heal-pulse': rowState(i) === 'healing' }]"
      >
        <code class="text-[11px] font-mono text-slate-700 flex-shrink-0" :class="{ 'line-through text-slate-400': rowState(i) === 'deleted' }">{{ r.col }}</code>
        <span class="text-[9px] text-slate-400 font-medium flex-shrink-0">{{ r.tag }}</span>

        <!-- soft-flag chips -->
        <div class="ml-auto flex items-center gap-1">
          <span
            class="w-4 h-4 rounded flex-center text-[8px] font-bold border transition-colors duration-300"
            :class="flags(i).e ? 'border-amber-300 bg-amber-100 text-amber-700' : 'border-gray-200 bg-gray-50 text-gray-300'"
            title="is_expired"
          >E</span>
          <span
            class="w-4 h-4 rounded flex-center text-[8px] font-bold border transition-colors duration-300"
            :class="flags(i).s ? 'border-violet-300 bg-violet-100 text-violet-700' : 'border-gray-200 bg-gray-50 text-gray-300'"
            title="is_stale"
          >S</span>
          <span
            class="w-4 h-4 rounded flex-center text-[8px] font-bold border transition-colors duration-300"
            :class="flags(i).d ? 'border-rose-300 bg-rose-100 text-rose-700' : 'border-gray-200 bg-gray-50 text-gray-300'"
            title="is_deleted"
          >D</span>
        </div>

        <!-- state icon -->
        <div class="w-4 flex-shrink-0 flex-center">
          <div v-if="rowState(i) === 'healing'" class="i-lucide-loader-circle text-amber-500 text-xs animate-spin" />
          <div v-else-if="rowState(i) === 'healed'" class="i-lucide-check text-emerald-500 text-xs" />
          <div v-else-if="rowState(i) === 'deleted'" class="i-lucide-trash-2 text-gray-400 text-xs" />
          <div v-else-if="rowState(i) === 'expired'" class="i-lucide-flag text-amber-500 text-xs" />
          <div v-else class="w-1.5 h-1.5 rounded-full bg-gray-200" />
        </div>
      </div>
    </div>

    <!-- footer: flag legend -->
    <div class="flex items-center flex-wrap gap-x-3 gap-y-1 px-3.5 py-2.5 border-t border-slate-100 text-[10px] text-slate-500">
      <div class="flex items-center gap-1"><span class="w-3.5 h-3.5 rounded flex-center text-[7px] font-bold border border-amber-300 bg-amber-100 text-amber-700">E</span>is_expired</div>
      <div class="flex items-center gap-1"><span class="w-3.5 h-3.5 rounded flex-center text-[7px] font-bold border border-violet-300 bg-violet-100 text-violet-700">S</span>is_stale</div>
      <div class="flex items-center gap-1"><span class="w-3.5 h-3.5 rounded flex-center text-[7px] font-bold border border-rose-300 bg-rose-100 text-rose-700">D</span>is_deleted</div>
      <span class="ml-auto px-2 py-0.5 rounded-full bg-slate-50 text-slate-500 border border-slate-200 font-mono">写即标脏 · 收尾重嵌</span>
    </div>
  </div>
</template>

<style scoped>
.heal-pulse {
  animation: heal-pulse 0.9s ease-in-out infinite;
}
@keyframes heal-pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(245, 158, 11, 0.35); }
  50% { box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.12); }
}
</style>
