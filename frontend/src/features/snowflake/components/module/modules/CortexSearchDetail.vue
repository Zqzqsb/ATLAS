<script setup lang="ts">
import { computed } from 'vue'
import ArchBox from '../../../../arch/components/module/diagram/ArchBox.vue'
import Connector from '../../../../arch/components/module/diagram/Connector.vue'
import InsightNotes from '../../../../arch/components/module/diagram/InsightNotes.vue'
import EvidenceChip from '../../../../arch/components/module/diagram/EvidenceChip.vue'
import { SOURCES } from '../../../model/sources'
import type { SnowFlowDef } from '../../../model/flows'
import { getSnowModule } from '../../../model/modules'

const props = defineProps<{ flow: SnowFlowDef; showNotes?: boolean }>()
const arch = computed(() => getSnowModule(props.flow.id)?.cortexSearch ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- Stage 1: motivation -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="blue" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-alert-triangle" title="Why Cortex Search" role="motivation" accent="blue" muted>
      <template #refs>
        <EvidenceChip :refs="arch.motivation.refs" :catalog="SOURCES" />
      </template>
      <ul class="space-y-1.5 pl-0 list-none">
        <li v-for="(p, i) in arch.motivation.points" :key="i" class="flex items-start gap-2">
          <div class="i-lucide-circle-dot text-blue-400 text-xs mt-1 flex-shrink-0" />
          <span class="text-[11px] text-gray-600 leading-snug">{{ p }}</span>
        </li>
      </ul>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- Stage 2: hybrid retrieval -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="blue" :items="arch.insights.hybrid" />
    </div>
    <div>
      <Connector label="hybrid retrieval" />
      <ArchBox icon="i-lucide-search" title="Cortex Search" role="向量 + 关键词 + rerank" accent="blue">
        <template #refs>
          <EvidenceChip :refs="['S6', 'S7']" :catalog="SOURCES" />
        </template>
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
          <div v-for="h in arch.hybrid" :key="h.name" class="rounded-lg border border-blue-200 bg-blue-50/40 px-2 py-1.5">
            <div class="flex items-center gap-1.5 mb-0.5">
              <code class="text-[11px] font-mono font-bold text-blue-700">{{ h.name }}</code>
              <EvidenceChip v-if="h.refs" class="ml-auto" :refs="h.refs" :catalog="SOURCES" size="xs" />
            </div>
            <span class="text-[11px] text-gray-500 leading-snug">{{ h.desc }}</span>
          </div>
        </div>
        <div class="mt-2.5 rounded-lg border border-dashed border-blue-200 bg-blue-50/30 px-2.5 py-1.5">
          <div class="flex items-center gap-1.5 mb-1">
            <div class="i-lucide-target text-blue-500 text-xs" />
            <span class="text-[11px] font-bold text-blue-700">Serves at runtime</span>
          </div>
          <ul class="space-y-1 pl-0 list-none">
            <li v-for="s in arch.serves" :key="s.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-blue-700 flex-shrink-0">{{ s.name }}</code>
              <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ s.desc }}</span>
              <EvidenceChip v-if="s.refs" :refs="s.refs" :catalog="SOURCES" size="xs" />
            </li>
          </ul>
        </div>
      </ArchBox>
    </div>
    <!-- right: yaml config -->
    <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
      <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
        <div class="i-lucide-file-code-2 text-blue-500 text-sm" />
        <span class="text-sm font-bold text-slate-800">cortex_search_service 配置</span>
        <EvidenceChip class="ml-auto" :refs="['S6']" :catalog="SOURCES" size="xs" />
      </div>
      <pre class="text-[10.5px] leading-relaxed font-mono text-gray-700 p-3 overflow-auto whitespace-pre">{{ arch.yamlConfig }}</pre>
    </div>

    <!-- Stage 3: integration -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="blue" :items="arch.insights.integrate" />
    </div>
    <div>
      <Connector label="dimension → search service" />
      <ArchBox icon="i-lucide-share-2" title="Integration with Semantic View" role="literal 召回外置" accent="blue">
        <template #refs>
          <EvidenceChip :refs="arch.integration.refs" :catalog="SOURCES" />
        </template>
        <ul class="space-y-1.5 pl-0 list-none">
          <li v-for="(p, i) in arch.integration.points" :key="i" class="flex items-start gap-2">
            <div class="i-lucide-arrow-right text-blue-400 text-xs mt-1 flex-shrink-0" />
            <span class="text-[11px] text-gray-600 leading-snug">{{ p }}</span>
          </li>
        </ul>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
