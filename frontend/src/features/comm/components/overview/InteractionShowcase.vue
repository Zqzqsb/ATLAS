<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { COMM_LAYERS, type InteractionScenario } from '../../model/comm'
import { ACCENTS } from '../../../arch/model/architecture'
import type { ArchNode } from '../../model/comm'
import InlineCode from '../module/InlineCode.vue'

/** The 4 entry-point nodes from the Interaction layer. */
const uxNodes = computed<ArchNode[]>(() => {
  const ux = COMM_LAYERS.find((l) => l.id === 'ux')
  return ux?.nodes ?? []
})

const activeIdx = ref(0)
const active = computed<ArchNode | undefined>(() => uxNodes.value[activeIdx.value])

let timer: number | null = null
function startAuto() {
  stopAuto()
  if (uxNodes.value.length < 2) return
  timer = window.setInterval(() => {
    activeIdx.value = (activeIdx.value + 1) % uxNodes.value.length
  }, 5200)
}
function stopAuto() {
  if (timer !== null) {
    window.clearInterval(timer)
    timer = null
  }
}
function select(i: number) {
  activeIdx.value = i
  startAuto()
}

onMounted(startAuto)
onUnmounted(stopAuto)
watch(uxNodes, startAuto)

/** Pretty accent ring for the active pill. */
function ringClass(n: ArchNode) {
  const a = ACCENTS[n.accent]
  return a.gradient
}
</script>

<template>
  <div
    v-if="active"
    class="relative rounded-2xl border border-slate-200 overflow-hidden shadow-md bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900"
  >
    <!-- subtle scan-line texture, gives it the "TV screen" vibe -->
    <div
      class="pointer-events-none absolute inset-0 opacity-[0.07]"
      style="background-image: repeating-linear-gradient(0deg, rgba(255,255,255,0.5) 0 1px, transparent 1px 3px);"
    />
    <!-- vignette -->
    <div class="pointer-events-none absolute inset-0" style="background: radial-gradient(ellipse at center, transparent 55%, rgba(0,0,0,0.55) 100%);" />

    <div class="relative grid grid-cols-12 gap-0">
      <!-- LEFT: meta panel (channel name, big icon, key idea) -->
      <div class="col-span-12 md:col-span-4 px-5 py-5 border-b md:border-b-0 md:border-r border-white/10 text-white">
        <div class="flex items-center gap-1.5 text-[10px] font-bold tracking-[0.18em] text-slate-300/80 uppercase">
          <span class="w-1.5 h-1.5 rounded-full bg-rose-400 animate-pulse" />
          Channel {{ String(activeIdx + 1).padStart(2, '0') }} / {{ uxNodes.length }}
        </div>

        <div class="mt-3 flex items-center gap-2.5">
          <div class="w-10 h-10 rounded-xl flex-center shadow-md" :class="ringClass(active)">
            <div :class="[active.icon, 'text-white text-lg']" />
          </div>
          <div class="min-w-0">
            <div class="text-[16px] font-extrabold leading-tight">{{ active.label }}</div>
            <div class="text-[11px] text-slate-300/80 leading-snug">{{ active.sublabel }}</div>
          </div>
        </div>

        <div class="mt-4 space-y-1.5">
          <div
            v-for="(n, i) in uxNodes"
            :key="n.id"
            class="flex items-center gap-2 px-2 py-1.5 rounded-md cursor-pointer transition-all"
            :class="i === activeIdx ? 'bg-white/10 ring-1 ring-white/20' : 'hover:bg-white/5'"
            @click="select(i)"
          >
            <div class="w-6 h-6 rounded-md flex-center flex-shrink-0" :class="i === activeIdx ? ringClass(n) : 'bg-white/5'">
              <div :class="[n.icon, i === activeIdx ? 'text-white text-[12px]' : 'text-slate-400 text-[12px]']" />
            </div>
            <span class="text-[12px] font-semibold" :class="i === activeIdx ? 'text-white' : 'text-slate-400'">
              {{ n.label }}
            </span>
            <span v-if="i === activeIdx" class="ml-auto text-[9px] font-bold tracking-wider text-emerald-300 uppercase">on air</span>
          </div>
        </div>
      </div>

      <!-- RIGHT: rotating mini-UI stage -->
      <div class="col-span-12 md:col-span-8 px-5 py-5 min-h-[340px]">
        <Transition
          enter-active-class="transition-all duration-500 ease-out"
          enter-from-class="opacity-0 translate-y-3 scale-[0.98]"
          enter-to-class="opacity-100 translate-y-0 scale-100"
          leave-active-class="transition-all duration-300 ease-in absolute"
          leave-from-class="opacity-100"
          leave-to-class="opacity-0"
          mode="out-in"
        >
          <div :key="active.id" class="h-full">
            <!-- Chat UI mock -->
            <div v-if="active.scenario?.kind === 'chat'" class="space-y-3 h-full flex flex-col">
              <div class="text-[10px] font-bold tracking-widest text-emerald-300/80 uppercase">Chat UI · 自然语言</div>
              <div class="flex justify-end">
                <div class="max-w-[85%] rounded-2xl rounded-tr-md bg-indigo-500/90 text-white text-[12.5px] leading-relaxed px-3.5 py-2 shadow-md">
                  {{ active.scenario.userMsg }}
                </div>
              </div>
              <div v-if="active.scenario.clarification" class="flex justify-start">
                <div class="max-w-[90%] rounded-2xl rounded-tl-md bg-amber-500/15 border border-amber-400/40 text-amber-100 text-[12px] leading-relaxed px-3.5 py-2">
                  <div class="text-[9.5px] font-bold text-amber-300/80 tracking-wider mb-0.5">⚡ 澄清回路 · 主动反问</div>
                  {{ active.scenario.clarification }}
                </div>
              </div>
              <div class="flex justify-start flex-1">
                <div class="w-full max-w-[95%] rounded-2xl rounded-tl-md bg-slate-700/70 border border-white/10 text-slate-100 text-[12px] leading-relaxed px-3.5 py-2.5 shadow-md">
                  <div class="text-[9.5px] font-bold text-emerald-300/80 tracking-wider mb-1">SQL · 一键执行 / 编辑</div>
                  <pre class="font-mono text-[11.5px] text-emerald-200/90 whitespace-pre-wrap leading-relaxed"><code>{{ active.scenario.assistantSql }}</code></pre>
                  <div class="mt-2 flex gap-1.5">
                    <button class="px-2 py-0.5 text-[10px] font-bold rounded bg-emerald-500/80 text-white hover:bg-emerald-400">▶ 执行</button>
                    <button class="px-2 py-0.5 text-[10px] font-semibold rounded bg-white/10 text-slate-200 hover:bg-white/20">复制</button>
                    <button class="px-2 py-0.5 text-[10px] font-semibold rounded bg-white/10 text-slate-200 hover:bg-white/20">改写</button>
                  </div>
                </div>
              </div>
            </div>

            <!-- MCP / tool-calling mock -->
            <div v-else-if="active.scenario?.kind === 'mcp'" class="space-y-3 h-full flex flex-col">
              <div class="text-[10px] font-bold tracking-widest text-emerald-300/80 uppercase">MCP / SDK · Agent 工具腰带</div>
              <div class="rounded-lg bg-slate-700/50 border border-white/10 px-3 py-2 text-[12px] text-slate-200 leading-relaxed">
                <span class="text-[9.5px] font-bold text-violet-300/80 tracking-wider block mb-1">LLM reasoning</span>
                <InlineCode :text="active.scenario.reasoning" />
              </div>
              <div class="flex-1 grid grid-cols-2 gap-2">
                <div
                  v-for="(t, i) in active.scenario.tools"
                  :key="i"
                  class="rounded-lg px-2.5 py-2 border transition-all"
                  :class="t.status === 'calling'
                    ? 'bg-violet-500/20 border-violet-400/60 shadow-[0_0_0_3px_rgba(139,92,246,0.15)]'
                    : t.status === 'done'
                      ? 'bg-emerald-500/10 border-emerald-400/30'
                      : 'bg-slate-700/30 border-white/10'"
                >
                  <div class="flex items-center gap-1.5">
                    <div :class="[t.icon, t.status === 'calling' ? 'text-violet-300 text-[13px] animate-pulse' : t.status === 'done' ? 'text-emerald-300 text-[13px]' : 'text-slate-400 text-[13px]']" />
                    <code class="font-mono text-[11.5px] text-slate-100 font-bold">{{ t.name }}</code>
                    <span class="ml-auto text-[9px] font-bold uppercase tracking-wider" :class="t.status === 'calling' ? 'text-violet-300' : t.status === 'done' ? 'text-emerald-300' : 'text-slate-500'">
                      {{ t.status === 'calling' ? 'calling…' : t.status }}
                    </span>
                  </div>
                  <div v-if="t.output" class="mt-1 text-[10.5px] text-slate-300/80 font-mono pl-5">→ {{ t.output }}</div>
                </div>
              </div>
            </div>

            <!-- IDE / CLI diff mock -->
            <div v-else-if="active.scenario?.kind === 'ide'" class="space-y-2 h-full flex flex-col">
              <div class="text-[10px] font-bold tracking-widest text-emerald-300/80 uppercase">IDE / CLI · Git Diff 评审</div>
              <div class="flex items-center gap-2 text-[11px] font-mono">
                <div class="i-lucide-file-code text-slate-300" />
                <span class="text-slate-200 font-bold">{{ active.scenario.filePath }}</span>
                <span class="ml-auto px-1.5 py-0.5 rounded bg-emerald-500/20 text-emerald-300 text-[10px] font-bold">+{{ active.scenario.newLines.length - 1 }}</span>
              </div>
              <div class="text-[11.5px] text-slate-300 italic px-1">{{ active.scenario.commitMsg }}</div>
              <div class="flex-1 rounded-lg overflow-hidden border border-white/10 bg-slate-950/60 font-mono text-[11.5px] leading-relaxed">
                <div class="grid grid-cols-[auto_1fr]">
                  <div class="px-2 py-0.5 text-slate-500 text-right select-none border-r border-white/5">-1</div>
                  <div class="px-2 py-0.5 bg-rose-500/15 text-rose-200/80 whitespace-pre">{{ active.scenario.oldLine }}</div>
                  <template v-for="(ln, i) in active.scenario.newLines" :key="i">
                    <div class="px-2 py-0.5 text-slate-500 text-right select-none border-r border-white/5">{{ i + 1 }}</div>
                    <div class="px-2 py-0.5 bg-emerald-500/15 text-emerald-200/90 whitespace-pre">{{ ln }}</div>
                  </template>
                </div>
              </div>
              <div class="flex items-center gap-1.5 text-[10.5px] text-slate-400">
                <div class="i-lucide-git-pull-request" />
                <span>PR 自动从 commit 草拟 · reviewer 只需 +1 / 评论</span>
              </div>
            </div>

            <!-- BI bar chart mock -->
            <div v-else-if="active.scenario?.kind === 'bi'" class="space-y-3 h-full flex flex-col">
              <div class="text-[10px] font-bold tracking-widest text-emerald-300/80 uppercase">BI / Notebook · 嵌入图表</div>
              <div class="text-[13px] text-slate-100 font-bold">{{ active.scenario.caption }}</div>
              <div class="flex-1 grid grid-cols-[auto_1fr] gap-3 items-end">
                <!-- y axis label -->
                <div class="text-[9.5px] text-slate-400 font-mono self-start rotate-180" style="writing-mode: vertical-rl;">{{ active.scenario.yLabel }}</div>
                <div class="space-y-2 self-stretch flex flex-col justify-end">
                  <div class="flex items-end justify-around gap-3 h-[180px]">
                    <div
                      v-for="(b, i) in active.scenario.bars"
                      :key="i"
                      class="flex-1 flex flex-col items-center gap-1 group"
                    >
                      <div class="text-[10px] font-bold font-mono" :class="b.highlight ? 'text-amber-300' : 'text-slate-400'">{{ b.value }}</div>
                      <div
                        class="w-full rounded-t-md transition-all shadow-sm"
                        :class="b.highlight
                          ? 'bg-gradient-to-t from-amber-500 to-amber-300 ring-1 ring-amber-200/60'
                          : 'bg-gradient-to-t from-sky-500/70 to-sky-300/80'"
                        :style="{ height: `${Math.min(100, (b.value / 600) * 100)}%` }"
                      />
                      <div class="text-[10.5px] text-slate-400 font-mono">{{ b.label }}</div>
                    </div>
                  </div>
                  <div class="text-center text-[9.5px] text-slate-500 font-mono border-t border-white/10 pt-1">{{ active.scenario.xLabel }}</div>
                </div>
              </div>
              <div class="text-[10.5px] text-slate-400 italic">🪄 业务用户在已有 BI 里点 "用自然语言解释这张图 / 加个对比维度" → 由 Context Layer 实时给一段 SQL 算出来。</div>
            </div>
          </div>
        </Transition>
      </div>
    </div>

    <!-- progress bar -->
    <div class="absolute bottom-0 left-0 right-0 h-0.5 bg-white/5 overflow-hidden">
      <div
        class="h-full bg-gradient-to-r from-emerald-400 to-cyan-300 origin-left"
        :key="activeIdx"
        style="animation: showcase-progress 5.2s linear infinite;"
      />
    </div>
  </div>
</template>

<style scoped>
@keyframes showcase-progress {
  from { transform: scaleX(0); }
  to   { transform: scaleX(1); }
}
</style>
