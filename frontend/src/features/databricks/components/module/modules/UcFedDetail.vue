<script setup lang="ts">
import { computed } from 'vue'
import ArchBox from '../../../../arch/components/module/diagram/ArchBox.vue'
import Connector from '../../../../arch/components/module/diagram/Connector.vue'
import InsightNotes from '../../../../arch/components/module/diagram/InsightNotes.vue'
import EvidenceChip from '../../../../arch/components/module/diagram/EvidenceChip.vue'
import { SOURCES } from '../../../model/sources'
import type { DbxFlowDef } from '../../../model/flows'
import { getDbxModule } from '../../../model/modules'

const props = defineProps<{ flow: DbxFlowDef; showNotes?: boolean }>()
const arch = computed(() => getDbxModule(props.flow.id)?.ucFed ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- Stage 1: namespace -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="indigo" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-folder-tree" title="UC Namespace" role="root" accent="indigo">
      <template #refs>
        <EvidenceChip :refs="arch.ucNamespace.refs" :catalog="SOURCES" />
      </template>
      <code class="block text-[11px] font-mono font-bold text-indigo-700 px-2 py-1 bg-indigo-50 border border-indigo-200 rounded mb-2">{{ arch.ucNamespace.fqn }}</code>
      <p class="text-[11px] text-gray-500 leading-snug">{{ arch.ucNamespace.note }}</p>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- Stage 2: valid sources for metric view -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="indigo" :items="arch.insights.namespace" />
    </div>
    <div>
      <Connector label="metric_view.source ∈" />
      <ArchBox icon="i-lucide-archive" title="Valid Source Objects" role="metric view source" accent="indigo">
        <template #refs>
          <EvidenceChip :refs="['S2', 'S5']" :catalog="SOURCES" />
        </template>
        <ul class="space-y-1.5 pl-0 list-none">
          <li v-for="s in arch.sources" :key="s.name" class="flex items-baseline gap-2">
            <code class="text-[11px] font-mono font-semibold text-indigo-700 flex-shrink-0">{{ s.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ s.desc }}</span>
            <EvidenceChip v-if="s.refs" :refs="s.refs" :catalog="SOURCES" size="xs" />
          </li>
        </ul>
      </ArchBox>
    </div>
    <!-- right: connector grid -->
    <div class="rounded-2xl border border-slate-200 bg-white px-3.5 py-3">
      <div class="flex items-center gap-1.5 mb-2.5">
        <div class="i-lucide-network text-indigo-500 text-sm" />
        <span class="text-xs font-bold text-slate-700">Lakehouse Federation Connectors</span>
        <EvidenceChip class="ml-auto" :refs="['S11']" :catalog="SOURCES" size="xs" />
      </div>
      <div class="space-y-1.5">
        <div v-for="g in arch.connectors" :key="g.group" class="rounded-lg border border-indigo-100 bg-indigo-50/30 px-2.5 py-1.5">
          <div class="flex items-center gap-1.5 mb-0.5">
            <div :class="[g.icon, 'text-indigo-500 text-sm']" />
            <span class="text-[11px] font-bold text-slate-700">{{ g.group }}</span>
          </div>
          <span class="text-[11px] text-gray-500 leading-snug font-mono">{{ g.items }}</span>
        </div>
      </div>
    </div>

    <!-- Stage 3: multi-source semantic layer -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="indigo" :items="arch.insights.federate" />
    </div>
    <div>
      <Connector label="metric view 跨源 join" />
      <ArchBox icon="i-lucide-globe" title="Multi-Source Semantic Layer" role="跨源统一" accent="indigo">
        <template #refs>
          <EvidenceChip :refs="arch.multiSource.refs" :catalog="SOURCES" />
        </template>
        <ul class="space-y-1.5 pl-0 list-none">
          <li v-for="(p, i) in arch.multiSource.points" :key="i" class="flex items-start gap-2">
            <div class="i-lucide-circle-dot text-indigo-400 text-xs mt-1 flex-shrink-0" />
            <span class="text-[11px] text-gray-600 leading-snug">{{ p }}</span>
          </li>
        </ul>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
