<script setup lang="ts">
import { computed } from 'vue'
import { ACCENTS } from '../../../../arch/model/architecture'
import ArchBox from '../../../../arch/components/module/diagram/ArchBox.vue'
import Connector from '../../../../arch/components/module/diagram/Connector.vue'
import PeekPanel from '../../../../arch/components/module/diagram/PeekPanel.vue'
import InsightNotes from '../../../../arch/components/module/diagram/InsightNotes.vue'
import type { KtxFlowDef } from '../../../model/ktx'
import { getKtxModule } from '../../../model/ktx'

const props = defineProps<{ flow: KtxFlowDef; showNotes?: boolean }>()
const arch = computed(() => getKtxModule(props.flow.id)?.search ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="blue" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-search" :title="arch.input.label" accent="slate" muted>
      <code class="text-[11px] font-mono text-gray-600">{{ arch.input.example }}</code>
      <div class="text-[11px] text-gray-500 mt-1">{{ arch.input.note }}</div>
    </ArchBox>
    <div class="hidden lg:block" />

    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="blue" :items="arch.insights.score" />
    </div>
    <div>
      <Connector label="3 scorers" />
      <ArchBox icon="i-lucide-tally-5" title="HybridSearchCore" role="三路打分" accent="blue">
        <div class="space-y-1.5 mb-2">
          <div
            v-for="s in arch.scorers"
            :key="s.name"
            class="rounded-lg border px-2.5 py-2"
            :class="ACCENTS[s.accent].surface"
          >
            <div class="flex items-center gap-1.5">
              <div :class="[s.icon, ACCENTS[s.accent].text, 'text-sm']" />
              <span class="text-xs font-bold text-gray-700">{{ s.name }}</span>
              <code class="ml-auto text-[9px] font-mono text-gray-400">{{ s.weight }}</code>
            </div>
            <div class="text-[10.5px] text-gray-500 mt-0.5">{{ s.desc }}</div>
          </div>
        </div>
        <div class="rounded-xl border border-violet-200 bg-violet-50/40 px-2.5 py-2">
          <div class="flex items-center gap-1.5 mb-1">
            <div class="i-lucide-merge text-violet-500 text-sm" />
            <span class="text-xs font-bold text-gray-700">{{ arch.fusion.name }}</span>
          </div>
          <code class="text-[10.5px] font-mono text-violet-700">{{ arch.fusion.algo }}</code>
          <div class="text-[10.5px] text-gray-500 mt-1">{{ arch.fusion.note }}</div>
        </div>
      </ArchBox>
    </div>
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-calculator text-violet-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">RRF 示例</span>
        </div>
        <pre class="text-[10.5px] leading-relaxed font-mono text-gray-600 p-3 whitespace-pre">{{ arch.fusion.example }}</pre>
      </div>
    </div>

    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="violet" :items="arch.insights.fuse" />
    </div>
    <div>
      <Connector label="fused results" />
      <ArchBox icon="i-lucide-layers" title="检索面 + 索引" accent="indigo">
        <PeekPanel label="暴露面（MCP / CLI）" icon="i-lucide-plug" :count="arch.surfaces.length" accent="blue" class="mb-2">
          <div class="space-y-1.5">
            <div v-for="s in arch.surfaces" :key="s.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-blue-600 flex-shrink-0">{{ s.name }}</code>
              <span class="text-[11px] text-gray-500">{{ s.desc }}</span>
            </div>
          </div>
        </PeekPanel>
        <PeekPanel label="SQLite 索引表" icon="i-lucide-database" :count="arch.index.length" accent="indigo">
          <div class="space-y-1.5">
            <div v-for="t in arch.index" :key="t.table" class="rounded-lg border border-indigo-100 bg-indigo-50/30 px-2 py-1.5">
              <code class="text-[11px] font-mono font-bold text-indigo-700">{{ t.table }}</code>
              <div class="text-[10px] text-gray-400 font-mono">{{ t.cols.join(' · ') }}</div>
              <div class="text-[10.5px] text-gray-500">{{ t.note }}</div>
            </div>
          </div>
        </PeekPanel>
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-blue-200 bg-blue-50/40 px-2.5 py-1.5 mt-2">
          <div class="i-lucide-spline text-blue-400 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-blue-700 leading-snug">
            {{ arch.embeddingProvider.name }} · {{ arch.embeddingProvider.dim }} — {{ arch.embeddingProvider.note }}
          </span>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
