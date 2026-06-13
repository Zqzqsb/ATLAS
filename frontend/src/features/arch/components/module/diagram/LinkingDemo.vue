<script setup lang="ts">
/**
 * Accelerated demo of the LargeScale (react) grounding timing design.
 * Self-contained looping animation, two stacked parts driven by one playhead:
 *   1. Recall funnel       — All tables → HNSW candidates → LLM refine (count collapse).
 *   2. Concurrency timeline — Agent (LinkAsync) and CoarseRetriever launch at T0 in
 *      parallel; retrieval hands schema to the Agent via a shared slot, so reasoning
 *      overlaps retrieval. End-to-end ≈ max(retrieval, reasoning), not the sum.
 * Mirrors backend/internal/grounding/adaptive_pipeline.go (groundLargeScaleReact).
 */
import { ref, computed, onMounted, onUnmounted } from 'vue'

// playhead 0..100 over one end-to-end pass; HANDOFF is where retrieval finishes.
const HANDOFF = 38 // retrieval_latency boundary (T1)
const progress = ref(0)

// ─── derived fills (clamped to [0,1]) ───
const clamp = (v: number) => Math.max(0, Math.min(1, v))
const retrievalFill = computed(() => clamp(progress.value / HANDOFF) * 100)
const agentPreFill = computed(() => clamp(progress.value / HANDOFF) * 100) // 0→T1 planning/warm-up
const agentReasonFill = computed(() => clamp((progress.value - HANDOFF) / (100 - HANDOFF)) * 100)

const handed = computed(() => progress.value >= HANDOFF)
const done = computed(() => progress.value >= 100)

// ─── funnel stages ───
// 0: all tables, 1: HNSW candidates (after handoff), 2: LLM refined (near end)
const funnelStage = computed(() => (progress.value < HANDOFF ? 0 : progress.value < 88 ? 1 : 2))
const steps = [
  { key: 0, icon: 'i-lucide-layout-grid', label: '全部表', count: 30, accent: 'slate' },
  { key: 1, icon: 'i-lucide-radar', label: 'HNSW 候选', count: 8, accent: 'blue' },
  { key: 2, icon: 'i-lucide-brain', label: 'LLM 精选', count: 3, accent: 'emerald' },
] as const

// ─── looping playhead ───
const STEP = 1.4
const TICK_MS = 28
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
      <div class="i-lucide-git-compare-arrows text-blue-500 text-sm" />
      <span class="text-sm font-bold text-slate-800">Schema Linking</span>
      <span class="text-[11px] text-slate-400 ml-1">加速演示 · 大库 react 模式</span>
      <span
        class="ml-auto px-2 py-0.5 rounded-full text-[11px] font-semibold transition-colors"
        :class="done ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-blue-50 text-blue-700 border border-blue-200'"
      >{{ done ? 'Linked' : 'Linking…' }}</span>
    </div>

    <!-- ── Part 1: recall funnel ── -->
    <div class="px-3.5 pt-3 pb-2">
      <div class="flex items-stretch gap-1.5">
        <template v-for="(s, i) in steps" :key="s.key">
          <div
            class="flex-1 rounded-xl border px-2 py-2 text-center transition-all duration-500"
            :class="[
              funnelStage >= s.key
                ? s.accent === 'slate' ? 'border-slate-300 bg-slate-50'
                  : s.accent === 'blue' ? 'border-blue-300 bg-blue-50'
                  : 'border-emerald-300 bg-emerald-50'
                : 'border-gray-100 bg-gray-50/40 opacity-50',
            ]"
          >
            <div class="flex items-center justify-center gap-1 mb-0.5">
              <div
                :class="[s.icon, 'text-[11px]',
                  s.accent === 'slate' ? 'text-slate-500' : s.accent === 'blue' ? 'text-blue-500' : 'text-emerald-500']"
              />
              <span class="text-[10px] font-medium text-gray-500">{{ s.label }}</span>
            </div>
            <div
              class="text-lg font-extrabold tabular-nums transition-colors duration-500"
              :class="funnelStage >= s.key
                ? s.accent === 'slate' ? 'text-slate-700' : s.accent === 'blue' ? 'text-blue-700' : 'text-emerald-700'
                : 'text-gray-300'"
            >{{ s.count }}</div>
          </div>
          <div v-if="i < steps.length - 1" class="flex items-center">
            <div
              class="i-lucide-chevron-right text-base transition-colors duration-300"
              :class="funnelStage > s.key ? 'text-gray-400' : 'text-gray-200'"
            />
          </div>
        </template>
      </div>
    </div>

    <!-- ── Part 2: concurrency timeline (Gantt) ── -->
    <div class="px-3.5 pt-1 pb-3">
      <div class="text-[11px] font-semibold text-slate-500 mb-1.5 flex items-center gap-1">
        <div class="i-lucide-clock text-slate-400" /> 并发时序 · 检索 ∥ 推理
      </div>

      <div class="relative">
        <!-- playhead -->
        <div
          class="absolute top-0 bottom-5 w-px bg-rose-400/70 z-10 pointer-events-none"
          :style="{ left: `calc(${progress}% )` }"
        >
          <div class="absolute -top-0.5 -left-[3px] w-[7px] h-[7px] rounded-full bg-rose-400" />
        </div>

        <!-- Agent track -->
        <div class="flex items-center gap-2 mb-1.5">
          <span class="w-14 flex-shrink-0 text-[10px] font-semibold text-violet-600 text-right">Agent</span>
          <div class="relative flex-1 h-6 rounded-md bg-gray-100 overflow-hidden">
            <!-- planning / warm-up zone (0 → T1) -->
            <div
              class="absolute inset-y-0 left-0 bg-violet-200/70 transition-[width] duration-100 ease-linear flex items-center"
              :style="{ width: `calc(${agentPreFill}% * ${HANDOFF} / 100)` }"
            >
              <span class="text-[9px] text-violet-700 font-medium pl-1.5 whitespace-nowrap">规划</span>
            </div>
            <!-- reasoning zone (T1 → T2) -->
            <div
              class="absolute inset-y-0 bg-violet-500 transition-[width] duration-100 ease-linear flex items-center"
              :style="{ left: HANDOFF + '%', width: `calc(${agentReasonFill}% * ${100 - HANDOFF} / 100)` }"
            >
              <span class="text-[9px] text-white font-semibold pl-1.5 whitespace-nowrap">LLM 精选推理</span>
            </div>
          </div>
        </div>

        <!-- Retrieval track -->
        <div class="flex items-center gap-2 mb-1">
          <span class="w-14 flex-shrink-0 text-[10px] font-semibold text-blue-600 text-right">Retrieval</span>
          <div class="relative flex-1 h-6 rounded-md bg-gray-100 overflow-hidden">
            <div
              class="absolute inset-y-0 left-0 bg-blue-500 transition-[width] duration-100 ease-linear flex items-center"
              :style="{ width: `calc(${retrievalFill}% * ${HANDOFF} / 100)` }"
            >
              <span class="text-[9px] text-white font-semibold pl-1.5 whitespace-nowrap">4 路 HNSW 召回</span>
            </div>
            <!-- handoff marker -->
            <div
              class="absolute inset-y-0 w-0.5 bg-amber-400 transition-opacity duration-300"
              :style="{ left: HANDOFF + '%' }"
              :class="handed ? 'opacity-100' : 'opacity-0'"
            />
          </div>
        </div>

        <!-- handoff label -->
        <div class="flex items-center gap-2">
          <span class="w-14 flex-shrink-0" />
          <div class="relative flex-1 h-4">
            <div
              class="absolute -translate-x-1/2 flex items-center gap-0.5 text-[9px] font-medium text-amber-600 transition-opacity duration-300"
              :style="{ left: HANDOFF + '%' }"
              :class="handed ? 'opacity-100' : 'opacity-30'"
            >
              <div class="i-lucide-corner-left-up text-[10px]" />schema → slot
            </div>
          </div>
        </div>
      </div>

      <!-- compare footer -->
      <div class="mt-2.5 pt-2 border-t border-slate-100 flex items-center gap-3 text-[10px]">
        <div class="flex items-center gap-1.5">
          <span class="w-2 h-2 rounded-sm bg-violet-500" />
          <span class="text-slate-500">并发 <b class="text-slate-700">≈ 1.0×</b></span>
        </div>
        <div class="flex items-center gap-1.5">
          <span class="w-2 h-2 rounded-sm bg-gray-300" />
          <span class="text-slate-400 line-through">串行 ≈ 1.4×</span>
        </div>
        <span class="ml-auto px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200 font-semibold">
          重叠省时 ~25%
        </span>
      </div>
    </div>
  </div>
</template>
