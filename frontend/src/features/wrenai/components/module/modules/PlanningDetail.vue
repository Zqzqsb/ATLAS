<script setup lang="ts">
import { computed } from 'vue'
import { ACCENTS } from '../../../../arch/model/architecture'
import ArchBox from '../../../../arch/components/module/diagram/ArchBox.vue'
import Connector from '../../../../arch/components/module/diagram/Connector.vue'
import PeekPanel from '../../../../arch/components/module/diagram/PeekPanel.vue'
import InsightNotes from '../../../../arch/components/module/diagram/InsightNotes.vue'
import type { WrenFlowDef } from '../../../model/wren'
import { getWrenModule } from '../../../model/wren'

const props = defineProps<{ flow: WrenFlowDef; showNotes?: boolean }>()
const arch = computed(() => getWrenModule(props.flow.id)?.planning ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- ════ Stage 1: Input (Modeled SQL) ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="slate" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-file-code-2" :title="arch.input.label" accent="slate" muted>
      <div class="text-xs text-gray-500 leading-snug mb-1.5">{{ arch.input.note }}</div>
      <code class="block px-2 py-1 rounded bg-gray-100 text-gray-600 font-mono text-[11px]">{{ arch.input.example }}</code>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- ════ Stage 2: SQL Planner (wren-core) ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="amber" :items="arch.insights.transform" />
    </div>
    <div>
      <Connector label="modeled SQL" />
      <ArchBox icon="i-lucide-cpu" title="SQL Planner" role="dry-plan" accent="amber">
        <!-- the three collaborating engines -->
        <div class="grid grid-cols-3 gap-1.5 mb-2.5">
          <div v-for="c in arch.collaborators" :key="c.name" class="rounded-xl border p-2" :class="ACCENTS[c.accent].surface">
            <div class="flex items-center gap-1 mb-0.5">
              <div :class="[c.icon, ACCENTS[c.accent].text, 'text-xs flex-shrink-0']" />
              <span class="text-[11px] font-bold text-gray-800 truncate">{{ c.name }}</span>
            </div>
            <div class="text-[10px] text-gray-500 leading-snug">{{ c.desc }}</div>
            <code class="text-[9px] font-mono mt-0.5 inline-block" :class="ACCENTS[c.accent].text">{{ c.lang }}</code>
          </div>
        </div>
        <!-- numbered transform pipeline -->
        <ol class="space-y-1 pl-0 list-none mb-2.5">
          <li v-for="(s, i) in arch.steps" :key="s.name" class="flex items-start gap-2 text-[11px] text-gray-600 leading-relaxed">
            <span class="w-3.5 h-3.5 rounded-full bg-amber-100 text-amber-700 flex-center text-[8px] font-bold flex-shrink-0 mt-0.5">{{ i + 1 }}</span>
            <span><code class="font-mono font-semibold text-amber-700">{{ s.name }}</code> · {{ s.desc }}</span>
          </li>
        </ol>
        <PeekPanel label="wren-core 展开什么" icon="i-lucide-cog" :count="arch.expands.length" accent="violet">
          <div class="space-y-1.5">
            <div v-for="e in arch.expands" :key="e.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-violet-600 flex-shrink-0">{{ e.name }}</code>
              <span class="text-[11px] text-gray-500 leading-snug">{{ e.desc }}</span>
            </div>
          </div>
        </PeekPanel>
      </ArchBox>
    </div>
    <!-- right: modeled SQL → expanded SQL -->
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-git-compare-arrows text-amber-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">wren-core 展开</span>
          <span class="text-[11px] text-slate-400 ml-auto">不连库 · dry-plan</span>
        </div>
        <pre class="text-[10.5px] leading-relaxed font-mono text-gray-700 p-3 overflow-auto whitespace-pre">{{ arch.sqlBefore }}</pre>
        <div class="flex items-center justify-center py-0.5 border-t border-slate-100">
          <div class="i-lucide-chevron-down text-amber-300 text-sm" />
        </div>
        <pre class="text-[10.5px] leading-relaxed font-mono text-gray-500 bg-amber-50/40 p-3 overflow-auto whitespace-pre">{{ arch.sqlAfter }}</pre>
      </div>
    </div>

    <!-- ════ Stage 3: Policy gate ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="rose" :items="arch.insights.policy" />
    </div>
    <div>
      <Connector label="expanded SQL" />
      <ArchBox icon="i-lucide-shield-check" title="Policy Checks" role="规划即执法" accent="rose">
        <div class="space-y-1.5">
          <div v-for="p in arch.policy" :key="p.name" class="flex items-baseline gap-2">
            <code class="text-[11px] font-mono font-semibold text-rose-600 flex-shrink-0">{{ p.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug">{{ p.desc }}</span>
          </div>
        </div>
      </ArchBox>
    </div>
    <!-- right: target dialects -->
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white px-3.5 py-3">
        <div class="flex items-center gap-1.5 mb-2">
          <div class="i-lucide-languages text-indigo-500 text-sm" />
          <span class="text-xs font-bold text-slate-700">转译目标方言</span>
          <span class="px-1.5 rounded-full text-[10px] font-bold ml-auto bg-indigo-50 text-indigo-700 border border-indigo-200">{{ arch.dialects.count }}</span>
        </div>
        <p class="text-[11px] text-gray-500 leading-relaxed font-mono">{{ arch.dialects.list }}</p>
      </div>
    </div>

    <!-- ════ Stage 4: Output ════ -->
    <div v-if="showNotes" class="hidden lg:block" />
    <div>
      <Connector label="transpile" />
      <ArchBox icon="i-lucide-check-check" :title="arch.output.label" accent="indigo">
        <div class="text-xs text-gray-500 leading-snug">{{ arch.output.note }}</div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
