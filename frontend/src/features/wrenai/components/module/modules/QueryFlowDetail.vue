<script setup lang="ts">
import { computed } from 'vue'
import { ACCENTS } from '../../../../arch/model/architecture'
import ArchBox from '../../../../arch/components/module/diagram/ArchBox.vue'
import Connector from '../../../../arch/components/module/diagram/Connector.vue'
import InsightNotes from '../../../../arch/components/module/diagram/InsightNotes.vue'
import type { WrenFlowDef } from '../../../model/wren'
import { getWrenModule } from '../../../model/wren'

const props = defineProps<{ flow: WrenFlowDef; showNotes?: boolean }>()
const arch = computed(() => getWrenModule(props.flow.id)?.query ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- ════ Stage 1: Input ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="slate" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-message-square" :title="arch.input.label" accent="slate" muted>
      <div class="flex items-center gap-2 text-xs text-gray-500">
        <code class="px-1.5 py-0.5 rounded bg-gray-100 text-gray-600 font-mono text-[11px]">{{ arch.input.example }}</code>
        <span>{{ arch.input.note }}</span>
      </div>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- ════ Stage 2: Retrieve (Memory) ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="blue" :items="arch.insights.retrieve" />
    </div>
    <div>
      <Connector />
      <ArchBox icon="i-lucide-database" :title="arch.retrieve.title" :role="arch.retrieve.role" accent="blue">
        <div class="space-y-1.5 mb-2">
          <div class="rounded-lg bg-white border border-blue-100 px-2 py-1.5">
            <code class="text-[11px] font-mono font-bold text-blue-700">{{ arch.retrieve.recall.cmd }}</code>
            <div class="text-[11px] text-gray-500 leading-snug">{{ arch.retrieve.recall.desc }}</div>
          </div>
          <div class="rounded-lg bg-white border border-blue-100 px-2 py-1.5">
            <code class="text-[11px] font-mono font-bold text-blue-700">{{ arch.retrieve.fetch.cmd }}</code>
            <div class="text-[11px] text-gray-500 leading-snug">{{ arch.retrieve.fetch.desc }}</div>
          </div>
          <div class="flex items-start gap-1.5 text-[11px] text-gray-500 leading-snug px-1">
            <div class="i-lucide-file-text text-emerald-500 text-xs mt-0.5 flex-shrink-0" />
            <span>{{ arch.retrieve.instructions }}</span>
          </div>
        </div>
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-blue-200 bg-blue-50/40 px-2.5 py-1.5">
          <div class="i-lucide-info text-blue-400 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-blue-700 leading-snug">{{ arch.retrieve.note }}</span>
        </div>
      </ArchBox>
    </div>
    <!-- right: correctness primitives legend -->
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-blocks text-violet-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">正确性原语</span>
          <span class="text-[11px] text-slate-400 ml-auto">Agent 自行编排</span>
        </div>
        <div class="p-2.5 space-y-1.5">
          <div v-for="p in arch.primitives" :key="p.name" class="flex items-center gap-2 rounded-lg border border-slate-100 bg-slate-50/50 px-2 py-1.5">
            <code class="text-[11px] font-mono font-bold text-violet-700 flex-shrink-0">{{ p.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug ml-auto">{{ p.desc }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- ════ Stage 3: Generate (external agent) ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="slate" :items="arch.insights.generate" />
    </div>
    <div>
      <Connector label="context + few-shot" />
      <ArchBox icon="i-lucide-bot" :title="arch.generate.title" :role="arch.generate.role" accent="slate" muted>
        <ul class="space-y-1 mb-2">
          <li v-for="(p, i) in arch.generate.points" :key="i" class="flex items-start gap-2 text-xs text-gray-600 leading-relaxed">
            <div class="i-lucide-dot text-gray-400 flex-shrink-0" />
            <span>{{ p }}</span>
          </li>
        </ul>
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-gray-300 bg-gray-50 px-2.5 py-1.5">
          <div class="i-lucide-scroll-text text-gray-400 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-gray-600 leading-snug">{{ arch.generate.note }}</span>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />

    <!-- ════ Stage 4: Plan (dry-plan) ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="amber" :items="arch.insights.plan" />
    </div>
    <div>
      <Connector label="modeled SQL" />
      <ArchBox icon="i-lucide-cpu" :title="arch.plan.title" :role="arch.plan.role" accent="amber">
        <ol class="space-y-1 pl-0 list-none">
          <li v-for="(s, i) in arch.plan.steps" :key="s.name" class="flex items-start gap-2 text-[11px] text-gray-600 leading-relaxed">
            <span class="w-3.5 h-3.5 rounded-full bg-amber-100 text-amber-700 flex-center text-[8px] font-bold flex-shrink-0 mt-0.5">{{ i + 1 }}</span>
            <span><code class="font-mono font-semibold text-amber-700">{{ s.name }}</code> · {{ s.desc }}</span>
          </li>
        </ol>
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-amber-200 bg-amber-50/40 px-2.5 py-1.5 mt-2">
          <div class="i-lucide-route text-amber-500 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-amber-700 leading-snug">{{ arch.plan.note }}</span>
        </div>
      </ArchBox>
    </div>
    <!-- right: modeled SQL → expanded SQL -->
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-git-compare-arrows text-amber-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">dry-plan 展开</span>
          <span class="text-[11px] text-slate-400 ml-auto">不连库</span>
        </div>
        <pre class="text-[10.5px] leading-relaxed font-mono text-gray-700 p-3 overflow-auto whitespace-pre">{{ arch.plan.sqlBefore }}</pre>
        <div class="flex items-center justify-center py-0.5 border-t border-slate-100">
          <div class="i-lucide-chevron-down text-amber-300 text-sm" />
        </div>
        <pre class="text-[10.5px] leading-relaxed font-mono text-gray-500 bg-amber-50/40 p-3 overflow-auto whitespace-pre">{{ arch.plan.sqlAfter }}</pre>
      </div>
    </div>

    <!-- ════ Stage 5: Execute ════ -->
    <div v-if="showNotes" class="hidden lg:block" />
    <div>
      <Connector label="executable SQL" />
      <ArchBox icon="i-lucide-play" :title="arch.execute.title" :role="arch.execute.role" accent="indigo">
        <ul class="space-y-1 mb-2">
          <li v-for="(s, i) in arch.execute.steps" :key="i" class="flex items-start gap-2 text-xs text-gray-600 leading-relaxed">
            <div class="i-lucide-check mt-0.5 flex-shrink-0 text-indigo-500" />
            <span>{{ s }}</span>
          </li>
        </ul>
        <div class="rounded-xl border border-indigo-200 bg-indigo-50/40 px-2.5 py-2 flex items-center gap-1.5">
          <div class="i-lucide-table text-indigo-600 text-sm" />
          <span class="text-xs font-bold text-gray-700">{{ arch.execute.output }}</span>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />

    <!-- ════ Stage 6: Repair / Store (closed loop) ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="emerald" :items="arch.insights.store" />
    </div>
    <div>
      <Connector label="WrenError → retry" />
      <ArchBox icon="i-lucide-refresh-cw" :title="arch.repair.title" accent="rose">
        <div class="space-y-1.5 mb-2">
          <div v-for="r in arch.repair.items" :key="r.name" class="flex items-baseline gap-2">
            <code class="text-[11px] font-mono font-semibold text-rose-600 flex-shrink-0">{{ r.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug">{{ r.desc }}</span>
          </div>
        </div>
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-rose-200 bg-rose-50/40 px-2.5 py-1.5 mb-2.5">
          <div class="i-lucide-blocks text-rose-400 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-rose-700 leading-snug">{{ arch.repair.note }}</span>
        </div>
        <!-- store (closed loop) -->
        <div class="rounded-xl border border-emerald-200 bg-emerald-50/40 px-2.5 py-2">
          <div class="flex items-center gap-1.5 mb-1">
            <div class="i-lucide-recycle text-emerald-600 text-sm" />
            <code class="text-[11px] font-mono font-bold text-emerald-700">{{ arch.store.cmd }}</code>
          </div>
          <span class="text-[11px] text-gray-500 leading-snug">{{ arch.store.note }}</span>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
