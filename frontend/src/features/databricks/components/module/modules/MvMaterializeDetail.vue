<script setup lang="ts">
import { computed } from 'vue'
import ArchBox from '../../../../arch/components/module/diagram/ArchBox.vue'
import Connector from '../../../../arch/components/module/diagram/Connector.vue'
import PeekPanel from '../../../../arch/components/module/diagram/PeekPanel.vue'
import InsightNotes from '../../../../arch/components/module/diagram/InsightNotes.vue'
import EvidenceChip from '../../../../arch/components/module/diagram/EvidenceChip.vue'
import { SOURCES } from '../../../model/sources'
import type { DbxFlowDef } from '../../../model/flows'
import { getDbxModule } from '../../../model/modules'

const props = defineProps<{ flow: DbxFlowDef; showNotes?: boolean }>()
const arch = computed(() => getDbxModule(props.flow.id)?.mvMaterialize ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- Stage 1: input -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="violet" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-shapes" :title="arch.input.label" role="logical metric view" accent="amber" muted>
      <template #refs>
        <EvidenceChip :refs="arch.input.refs" :catalog="SOURCES" />
      </template>
      <p class="text-[11px] text-gray-500 leading-snug">{{ arch.input.note }}</p>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- Stage 2: setup MV -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="violet" :items="arch.insights.materialize" />
    </div>
    <div>
      <Connector label="CREATE MATERIALIZED" />
      <ArchBox icon="i-lucide-snowflake" title="Materialize Setup" role="增量物化" accent="violet">
        <template #refs>
          <EvidenceChip :refs="['S2', 'S14']" :catalog="SOURCES" />
        </template>
        <ul class="space-y-1.5 pl-0 list-none">
          <li v-for="s in arch.setup" :key="s.name" class="flex items-baseline gap-2">
            <code class="text-[11px] font-mono font-semibold text-violet-700 flex-shrink-0">{{ s.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ s.desc }}</span>
            <EvidenceChip v-if="s.refs" :refs="s.refs" :catalog="SOURCES" size="xs" />
          </li>
        </ul>
        <div class="grid grid-cols-2 gap-2 mt-2">
          <PeekPanel label="refresh modes" icon="i-lucide-refresh-ccw" :count="arch.refresh.length" accent="violet">
            <ul class="space-y-1 pl-0 list-none">
              <li v-for="r in arch.refresh" :key="r.name" class="flex items-baseline gap-2">
                <code class="text-[11px] font-mono font-semibold text-violet-700 flex-shrink-0">{{ r.name }}</code>
                <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ r.desc }}</span>
                <EvidenceChip v-if="r.refs" :refs="r.refs" :catalog="SOURCES" size="xs" />
              </li>
            </ul>
          </PeekPanel>
          <PeekPanel label="cardinality / rely" icon="i-lucide-fingerprint" :count="arch.cardinalityHints.length" accent="amber">
            <ul class="space-y-1 pl-0 list-none">
              <li v-for="h in arch.cardinalityHints" :key="h.name" class="flex items-baseline gap-2">
                <code class="text-[11px] font-mono font-semibold text-amber-700 flex-shrink-0">{{ h.name }}</code>
                <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ h.desc }}</span>
                <EvidenceChip v-if="h.refs" :refs="h.refs" :catalog="SOURCES" size="xs" />
              </li>
            </ul>
          </PeekPanel>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />

    <!-- Stage 3: rewrite -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="violet" :items="arch.insights.rewrite" />
    </div>
    <div>
      <Connector label="optimizer transparently rewrites" />
      <ArchBox icon="i-lucide-replace" title="Auto Query Rewrite" role="命中即跳聚合" accent="violet">
        <template #refs>
          <EvidenceChip :refs="arch.rewrite.refs" :catalog="SOURCES" />
        </template>
        <ul class="space-y-1.5 pl-0 list-none">
          <li v-for="(p, i) in arch.rewrite.points" :key="i" class="flex items-start gap-2">
            <div class="i-lucide-circle-dot text-violet-400 text-xs mt-1 flex-shrink-0" />
            <span class="text-[11px] text-gray-600 leading-snug">{{ p }}</span>
          </li>
        </ul>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
