<script setup lang="ts">
import { computed } from 'vue'
import { ACCENTS } from '../../../model/architecture'
import type { FlowDef } from '../../../model/flows'
import { getModule } from '../../../model/modules'
import ArchBox from '../diagram/ArchBox.vue'
import Connector from '../diagram/Connector.vue'
import PeekPanel from '../diagram/PeekPanel.vue'
import InsightNotes from '../diagram/InsightNotes.vue'

const props = defineProps<{ flow: FlowDef; showNotes?: boolean }>()
const arch = computed(() => getModule(props.flow.id)?.kernel ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)

// per-transcript-kind presentation
const KIND = {
  thought: { label: 'Thought', cls: 'text-violet-600', code: false },
  action: { label: 'Action', cls: 'text-blue-600', code: true },
  input: { label: 'Action Input', cls: 'text-gray-500', code: true },
  observation: { label: 'Observation', cls: 'text-emerald-600', code: false },
  final: { label: 'Final Answer', cls: 'text-amber-600', code: false },
} as const
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- ════ Stage 1: Scenario Config ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="violet" :items="arch.insights.scenarios" />
    </div>
    <ArchBox icon="i-lucide-sliders-horizontal" :title="arch.scenarios.title" :role="arch.scenarios.role" accent="slate" muted>
      <div class="space-y-1 mb-2.5">
        <div v-for="c in arch.scenarios.config" :key="c.name" class="flex items-baseline gap-2">
          <code class="text-[11px] font-mono font-semibold text-slate-600 flex-shrink-0">{{ c.name }}</code>
          <span class="text-[11px] text-gray-500 leading-snug">{{ c.desc }}</span>
        </div>
      </div>
      <PeekPanel label="场景 Builder × 6" icon="i-lucide-list-tree" :count="arch.scenarios.list.length" accent="violet">
        <div class="space-y-1.5">
          <div v-for="s in arch.scenarios.list" :key="s.name" class="rounded-lg border border-slate-100 bg-slate-50/60 px-2 py-1.5">
            <div class="flex items-center gap-1.5">
              <code class="text-[11px] font-mono font-bold text-violet-700">{{ s.name }}</code>
              <span class="ml-auto text-[10px] font-mono text-gray-400">{{ s.budget }}</span>
            </div>
            <div class="text-[10.5px] text-gray-500 leading-snug mt-0.5">{{ s.tools }}</div>
          </div>
        </div>
      </PeekPanel>
      <div class="flex items-start gap-2 rounded-lg border border-dashed border-slate-200 bg-slate-50/60 px-2.5 py-1.5 mt-2">
        <div class="i-lucide-info text-slate-400 text-xs mt-0.5 flex-shrink-0" />
        <span class="text-[11px] text-slate-600 leading-snug">{{ arch.scenarios.note }}</span>
      </div>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- ════ Stage 2: ReAct Engine loop ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="violet" :items="arch.insights.engine" />
    </div>
    <div>
      <Connector label="react.New(llm, EngineConfig)" />
      <ArchBox icon="i-lucide-repeat" :title="arch.engine.title" :role="arch.engine.role" accent="violet">
        <div class="text-[10px] text-gray-400 font-mono mb-2">{{ arch.engine.base }}</div>
        <!-- the loop phases -->
        <div class="grid grid-cols-3 gap-1.5 mb-2.5">
          <div v-for="(p, i) in arch.engine.loop" :key="p.name" class="rounded-xl border border-violet-100 bg-violet-50/40 p-2">
            <div class="flex items-center gap-1 mb-0.5">
              <span class="w-3.5 h-3.5 rounded-full bg-violet-100 text-violet-700 flex-center text-[8px] font-bold flex-shrink-0">{{ i + 1 }}</span>
              <span class="text-[11px] font-bold text-violet-700">{{ p.name }}</span>
            </div>
            <div class="text-[10px] text-gray-500 leading-snug">{{ p.desc }}</div>
          </div>
        </div>
        <div class="flex items-center justify-center gap-1.5 text-[10px] text-violet-400 mb-2">
          <div class="i-lucide-rotate-cw text-xs" />
          <span>循环直到 Final Answer 或迭代上限</span>
        </div>
        <div class="space-y-1.5">
          <div class="flex items-start gap-2 rounded-lg border border-dashed border-violet-200 bg-violet-50/40 px-2.5 py-1.5">
            <div class="i-lucide-type text-violet-400 text-xs mt-0.5 flex-shrink-0" />
            <span class="text-[11px] text-violet-700 leading-snug">{{ arch.engine.format }}</span>
          </div>
          <div class="flex items-start gap-2 rounded-lg border border-dashed border-violet-200 bg-violet-50/40 px-2.5 py-1.5">
            <div class="i-lucide-life-buoy text-violet-400 text-xs mt-0.5 flex-shrink-0" />
            <span class="text-[11px] text-violet-700 leading-snug">{{ arch.engine.parser }}</span>
          </div>
          <div class="flex items-start gap-2 rounded-lg border border-dashed border-violet-200 bg-violet-50/40 px-2.5 py-1.5">
            <div class="i-lucide-gauge text-violet-400 text-xs mt-0.5 flex-shrink-0" />
            <span class="text-[11px] text-violet-700 leading-snug">{{ arch.engine.budget }}</span>
          </div>
        </div>
      </ArchBox>
    </div>
    <!-- right: sample transcript -->
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-scroll text-violet-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">单轮 Transcript</span>
          <span class="text-[11px] text-slate-400 ml-auto">inference 场景</span>
        </div>
        <div class="p-3 space-y-1.5 bg-gray-900/95">
          <div v-for="(t, i) in arch.transcript" :key="i" class="flex items-baseline gap-2 text-[10.5px] leading-relaxed">
            <span class="font-mono font-bold flex-shrink-0 w-24 text-right" :class="KIND[t.kind].cls">{{ KIND[t.kind].label }}</span>
            <span class="font-mono" :class="KIND[t.kind].code ? 'text-cyan-300' : 'text-gray-200'">{{ t.text }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- ════ Stage 3: Tool Belt + LLM/output ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="amber" :items="arch.insights.tools" />
    </div>
    <div>
      <Connector label="Action → tools.Tool.Call()" />
      <ArchBox icon="i-lucide-wrench" title="Tool Belt" role="12 工具 · 4 组" accent="violet">
        <div class="space-y-2">
          <div v-for="g in arch.toolGroups" :key="g.name" class="rounded-xl border p-2.5" :class="ACCENTS[g.accent].surface">
            <div class="flex items-center gap-1.5 mb-1.5">
              <div :class="[g.icon, ACCENTS[g.accent].text, 'text-sm flex-shrink-0']" />
              <span class="text-xs font-bold text-gray-700">{{ g.name }}</span>
              <span class="px-1.5 rounded-full text-[10px] font-bold ml-auto" :class="ACCENTS[g.accent].chip">{{ g.items.length }}</span>
            </div>
            <div class="space-y-1">
              <div v-for="t in g.items" :key="t.name" class="flex items-baseline gap-2">
                <code class="text-[11px] font-mono font-semibold flex-shrink-0" :class="ACCENTS[g.accent].text">{{ t.name }}</code>
                <span class="text-[11px] text-gray-500 leading-snug">{{ t.desc }}</span>
              </div>
            </div>
          </div>
        </div>
      </ArchBox>
    </div>
    <!-- right: LLM provider + output -->
    <div class="space-y-3">
      <div class="rounded-2xl border border-slate-200 bg-white px-3.5 py-3">
        <div class="flex items-center gap-1.5 mb-1.5">
          <div class="i-lucide-brain text-slate-500 text-sm" />
          <span class="text-xs font-bold text-slate-700">LLM Provider</span>
          <code class="ml-auto px-1.5 rounded text-[10px] font-mono font-bold bg-slate-100 text-slate-600 border border-slate-200">{{ arch.llm.encoding }}</code>
        </div>
        <div class="text-[11px] text-gray-500 leading-snug">{{ arch.llm.provider }}</div>
        <div class="text-[10.5px] text-gray-400 leading-snug mt-0.5">{{ arch.llm.note }}</div>
      </div>
      <div class="rounded-2xl border border-violet-200 bg-violet-50/40 px-3.5 py-3">
        <div class="flex items-center gap-1.5 mb-1.5">
          <div class="i-lucide-arrow-down-to-line text-violet-600 text-sm" />
          <span class="text-xs font-bold text-gray-700">OUTPUT · {{ arch.output.label }}</span>
        </div>
        <div class="flex flex-wrap gap-1">
          <span
            v-for="p in arch.output.parts"
            :key="p"
            class="px-1.5 py-0.5 rounded-md bg-white border border-violet-200 text-[11px] text-violet-700"
          >{{ p }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
