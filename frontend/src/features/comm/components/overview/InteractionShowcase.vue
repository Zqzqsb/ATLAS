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

/** Auto-rotate disabled — user picks a channel explicitly. Kept as a
 *  no-op so a future toggle ("press space to autoplay") is a one-line
 *  change. */
let timer: number | null = null
function startAuto() {
  stopAuto()
  // intentionally empty: no auto-rotation
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
        <div class="flex items-center gap-1.5 text-[10px] font-bold tracking-[0.18em] text-slate-400/80 uppercase">
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-400" />
          {{ String(activeIdx + 1).padStart(2, '0') }} / {{ uxNodes.length }} · 入口形态
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
            <span v-if="i === activeIdx" class="ml-auto text-[9px] font-bold tracking-wider text-slate-500 uppercase">active</span>
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

            <!-- IDE / CLI · VSCode 编辑器（左）+ GitHub PR（右）双栏 -->
            <div v-else-if="active.scenario?.kind === 'ide'" class="space-y-2 h-full flex flex-col">
              <div class="text-[10px] font-bold tracking-widest text-emerald-300/80 uppercase">IDE / CLI · 编辑器 + PR 评审</div>

              <div class="flex-1 grid grid-cols-2 gap-2 min-h-0">
                <!-- LEFT: VSCode-style editor + AI chat side panel -->
                <div class="rounded-lg border border-white/10 bg-slate-950/70 overflow-hidden flex flex-col min-h-0">
                  <!-- editor tab bar -->
                  <div class="flex items-center gap-1.5 px-2 py-1.5 border-b border-white/5 bg-slate-900/80 text-[10px]">
                    <div class="i-lucide-file-code text-slate-400 text-[12px]" />
                    <span class="font-mono font-bold text-slate-200">{{ active.scenario.left.filePath }}</span>
                    <span class="ml-auto flex items-center gap-1 text-slate-500">
                      <span class="w-1.5 h-1.5 rounded-full bg-amber-400" /> unsaved
                    </span>
                  </div>
                  <!-- editor body + chat side panel -->
                  <div class="flex-1 grid grid-cols-[1fr_auto] min-h-0">
                    <!-- code lines -->
                    <div class="font-mono text-[10.5px] leading-[1.55] py-1 overflow-y-auto">
                      <div
                        v-for="(ln, i) in active.scenario.left.lines"
                        :key="i"
                        class="grid grid-cols-[28px_1fr] hover:bg-white/[0.02]"
                        :class="{
                          'bg-emerald-500/10 border-l-2 border-emerald-400': ln.kind === 'new-active',
                          'bg-emerald-500/5 border-l-2 border-emerald-500/40': ln.kind === 'new',
                        }"
                      >
                        <div class="text-right pr-1.5 text-slate-500 select-none">{{ i + 1 }}</div>
                        <div
                          class="px-1.5 whitespace-pre"
                          :class="{
                            'text-slate-300': ln.kind === 'kept' || ln.kind === 'margin',
                            'text-emerald-200': ln.kind === 'new' || ln.kind === 'new-active',
                          }"
                        >{{ ln.code || ' ' }}</div>
                      </div>
                    </div>
                    <!-- right-side AI chat panel -->
                    <div class="w-[140px] border-l border-white/10 bg-slate-900/40 flex flex-col text-[10px]">
                      <div class="px-2 py-1.5 border-b border-white/5 flex items-center gap-1 text-violet-300 font-bold">
                        <div class="i-lucide-sparkles text-[11px]" /> Copilot
                      </div>
                      <div class="flex-1 px-2 py-1.5 space-y-1.5 overflow-y-auto">
                        <div class="rounded-md bg-indigo-500/15 border border-indigo-400/30 text-indigo-100 px-1.5 py-1 leading-snug">
                          <div class="text-[8.5px] font-bold text-indigo-300/80 mb-0.5">@me</div>
                          {{ active.scenario.left.chat.userPrompt }}
                        </div>
                        <div class="rounded-md bg-violet-500/10 border border-violet-400/20 text-violet-100 px-1.5 py-1 leading-snug">
                          <div class="text-[8.5px] font-bold text-violet-300/80 mb-0.5">assistant</div>
                          {{ active.scenario.left.chat.aiReply }}
                        </div>
                      </div>
                      <div class="px-2 py-1.5 border-t border-white/5 text-slate-500 text-center font-bold tracking-wider">
                        ↑ submit · ⌘↩
                      </div>
                    </div>
                  </div>
                </div>

                <!-- RIGHT: GitHub-style PR page -->
                <div class="rounded-lg border border-white/10 bg-slate-950/70 overflow-hidden flex flex-col min-h-0">
                  <!-- PR title + meta -->
                  <div class="px-2.5 py-1.5 border-b border-white/5">
                    <div class="flex items-center gap-1.5 text-[9.5px] text-slate-500 font-bold tracking-wider uppercase">
                      <div class="i-lucide-git-pull-request text-emerald-400" /> Pull Request
                    </div>
                    <div class="text-[12px] font-extrabold text-slate-100 mt-0.5 leading-tight">{{ active.scenario.right.title }}</div>
                    <div class="flex items-center gap-1.5 mt-1 text-[10px]">
                      <span class="px-1.5 py-0.5 rounded bg-emerald-500/15 text-emerald-300 font-mono font-bold">+{{ active.scenario.right.additions }}</span>
                      <span class="px-1.5 py-0.5 rounded bg-rose-500/15 text-rose-300 font-mono font-bold">−{{ active.scenario.right.deletions }}</span>
                      <span
                        class="px-1.5 py-0.5 rounded font-mono font-bold"
                        :class="{
                          'bg-emerald-500/20 text-emerald-300': active.scenario.right.ciState === 'pass',
                          'bg-amber-500/20 text-amber-300': active.scenario.right.ciState === 'pending',
                          'bg-rose-500/20 text-rose-300':    active.scenario.right.ciState === 'fail',
                        }"
                      >
                        <span v-if="active.scenario.right.ciState === 'pass'">✓</span>
                        <span v-else-if="active.scenario.right.ciState === 'pending'">●</span>
                        <span v-else>✗</span>
                        {{ active.scenario.right.ciLabel.replace(/^[✓●✗]\s*/, '') }}
                      </span>
                      <span class="ml-auto flex items-center -space-x-1">
                        <span
                          v-for="r in active.scenario.right.reviewers"
                          :key="r"
                          class="w-4 h-4 rounded-full bg-gradient-to-br from-cyan-500 to-blue-600 text-white text-[8px] font-extrabold flex-center ring-2 ring-slate-950"
                          :title="`@${r}`"
                        >{{ r[0] }}</span>
                      </span>
                    </div>
                  </div>
                  <!-- diff body -->
                  <div class="flex-1 font-mono text-[10px] leading-[1.5] overflow-y-auto">
                    <div class="px-2 py-1 text-[9px] font-bold text-slate-500 border-b border-white/5 sticky top-0 bg-slate-950/95 backdrop-blur">
                      {{ active.scenario.right.patch.filePath }}
                    </div>
                    <div class="grid grid-cols-[28px_1fr]">
                      <template v-for="(line, i) in active.scenario.right.patch.oldBlock" :key="`o-${i}`">
                        <div class="text-right pr-1.5 text-slate-500 select-none bg-rose-500/5">{{ i + 1 }}</div>
                        <div class="px-1.5 bg-rose-500/10 text-rose-200/80 whitespace-pre">{{ line }}</div>
                      </template>
                      <template v-for="(line, i) in active.scenario.right.patch.newBlock" :key="`n-${i}`">
                        <div class="text-right pr-1.5 text-slate-500 select-none bg-emerald-500/5">
                          {{ active.scenario.right.patch.oldBlock.length + i + 1 }}
                        </div>
                        <div
                          class="px-1.5 whitespace-pre"
                          :class="line ? 'bg-emerald-500/10 text-emerald-200/90' : 'bg-emerald-500/5'"
                        >{{ line || ' ' }}</div>
                      </template>
                    </div>
                  </div>
                </div>
              </div>

              <!-- bottom caption -->
              <div class="flex items-center gap-1.5 text-[10.5px] text-slate-400 px-1">
                <div class="i-lucide-arrow-left-right text-slate-500" />
                <span>左：开发者在 IDE 里让 AI 写语义层 · 右：PR 评审页（diff + CI + reviewers）</span>
              </div>
            </div>

            <!-- BI / Notebook · 看板（左）+ Ask-this-chart NL 面板（右）双栏 -->
            <div v-else-if="active.scenario?.kind === 'bi'" class="space-y-2 h-full flex flex-col">
              <div class="text-[10px] font-bold tracking-widest text-emerald-300/80 uppercase">BI / Notebook · 在图表旁边问</div>

              <div class="flex-1 grid grid-cols-[1.05fr_1fr] gap-2 min-h-0">
                <!-- LEFT: static BI dashboard tile -->
                <div class="rounded-lg border border-white/10 bg-slate-900/40 overflow-hidden flex flex-col min-h-0">
                  <div class="px-2.5 py-1.5 border-b border-white/5 flex items-center gap-1.5">
                    <div class="i-lucide-bar-chart-3 text-[12px] text-amber-300" />
                    <div class="text-[12px] font-extrabold text-slate-100 leading-tight">{{ active.scenario.left.title }}</div>
                    <span class="ml-auto text-[8.5px] font-bold tracking-wider text-slate-500 uppercase">live</span>
                  </div>
                  <div class="text-[9.5px] text-slate-500 px-2.5 pb-1.5">{{ active.scenario.left.subtitle }}</div>
                  <div class="flex-1 px-3 pb-2 flex gap-2 items-end min-h-0">
                    <div class="text-[9px] text-slate-400 font-mono self-start rotate-180" style="writing-mode: vertical-rl;">{{ active.scenario.left.yLabel }}</div>
                    <div class="flex-1 flex flex-col justify-end min-h-0">
                      <!-- fixed-height bar lane (140px) so the bars have a real
                           pixel parent for their height: X% to resolve against. -->
                      <div class="flex items-end justify-around gap-3" style="height: 140px;">
                        <div
                          v-for="(b, i) in active.scenario.left.bars"
                          :key="i"
                          class="flex-1 flex flex-col items-center justify-end gap-1 h-full group"
                        >
                          <div class="text-[9.5px] font-bold font-mono" :class="b.highlight ? 'text-amber-300' : 'text-slate-400'">{{ b.value }}</div>
                          <div
                            class="w-full rounded-t-md transition-all shadow-sm flex-shrink-0"
                            :class="b.highlight
                              ? 'bg-gradient-to-t from-amber-500 to-amber-300 ring-1 ring-amber-200/60'
                              : 'bg-gradient-to-t from-sky-500/70 to-sky-300/80'"
                            :style="{ height: `${Math.min(100, (b.value / 600) * 100)}%`, minHeight: '4px' }"
                          />
                          <div class="text-[9.5px] text-slate-400 font-mono">{{ b.label }}</div>
                        </div>
                      </div>
                      <div class="text-center text-[9px] text-slate-500 font-mono border-t border-white/10 pt-1 mt-1">{{ active.scenario.left.xLabel }}</div>
                    </div>
                  </div>
                </div>

                <!-- RIGHT: "Ask this chart" NL side panel -->
                <div class="rounded-lg border border-white/10 bg-slate-950/70 overflow-hidden flex flex-col min-h-0">
                  <div class="px-2.5 py-1.5 border-b border-white/5 flex items-center gap-1.5">
                    <div class="i-lucide-message-circle-question text-amber-300 text-[12px]" />
                    <span class="text-[11.5px] font-extrabold text-slate-100">Ask this chart</span>
                    <span class="ml-auto text-[8.5px] font-bold tracking-wider text-slate-500 uppercase">NL → SQL</span>
                  </div>

                  <!-- suggested chips -->
                  <div class="px-2.5 py-1.5 flex flex-wrap gap-1 border-b border-white/5">
                    <span
                      v-for="(s, i) in active.scenario.right.suggestions"
                      :key="i"
                      class="px-1.5 py-0.5 rounded-full text-[9.5px] font-semibold bg-amber-500/10 text-amber-200 border border-amber-400/30 cursor-pointer hover:bg-amber-500/20"
                    >{{ s }}</span>
                  </div>

                  <!-- user prompt input -->
                  <div class="px-2.5 py-1.5 border-b border-white/5">
                    <div class="flex items-center gap-1.5 rounded-md bg-slate-800/70 border border-white/10 px-2 py-1.5">
                      <div class="i-lucide-sparkles text-[11px] text-violet-300" />
                      <span class="text-[10.5px] text-slate-200 font-medium">{{ active.scenario.right.userPrompt }}</span>
                      <div class="ml-auto i-lucide-corner-down-left text-[10px] text-slate-500" />
                    </div>
                  </div>

                  <!-- generated SQL -->
                  <div class="px-2.5 py-1.5 border-b border-white/5">
                    <div class="flex items-center gap-1 text-[8.5px] font-bold text-emerald-300/80 tracking-wider uppercase mb-0.5">
                      <div class="i-lucide-code-2 text-[10px]" /> Generated SQL
                    </div>
                    <pre class="rounded bg-slate-900/80 border border-white/5 px-2 py-1.5 text-[9.5px] leading-[1.5] font-mono text-emerald-200/90 overflow-x-auto whitespace-pre"><code>{{ active.scenario.right.sql }}</code></pre>
                  </div>

                  <!-- narrative answer -->
                  <div class="flex-1 px-2.5 py-1.5 overflow-y-auto">
                    <div class="flex items-center gap-1 text-[8.5px] font-bold text-amber-300/80 tracking-wider uppercase mb-0.5">
                      <div class="i-lucide-sparkles text-[10px]" /> Answer
                    </div>
                    <p class="text-[10.5px] text-slate-200 leading-relaxed">{{ active.scenario.right.answer }}</p>
                  </div>
                </div>
              </div>

              <!-- bottom caption -->
              <div class="flex items-center gap-1.5 text-[10.5px] text-slate-400 px-1">
                <div class="i-lucide-arrow-left-right text-slate-500" />
                <span>左：已有 BI 看板 · 右：图表旁 NL 输入框（建议 prompt + 生成 SQL + 自然语言结论）</span>
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </div>

    <!-- (no auto-progress bar — auto-rotate is disabled) -->
  </div>
</template>
