<script setup lang="ts">
import { computed } from 'vue'
import { ACCENTS } from '../../../../arch/model/architecture'
import ArchBox from '../../../../arch/components/module/diagram/ArchBox.vue'
import Connector from '../../../../arch/components/module/diagram/Connector.vue'
import PeekPanel from '../../../../arch/components/module/diagram/PeekPanel.vue'
import InsightNotes from '../../../../arch/components/module/diagram/InsightNotes.vue'
import EvidenceChip from '../../../../arch/components/module/diagram/EvidenceChip.vue'
import { SOURCES } from '../../../model/sources'
import type { DbxFlowDef } from '../../../model/flows'
import { getDbxModule } from '../../../model/modules'

const props = defineProps<{ flow: DbxFlowDef; showNotes?: boolean }>()
const arch = computed(() => getDbxModule(props.flow.id)?.metricView ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- Stage 1: Source / DDL -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="amber" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-file-code-2" :title="arch.source.label" role="DDL · YAML" accent="amber">
      <template #refs>
        <EvidenceChip :refs="arch.source.refs" :catalog="SOURCES" />
      </template>
      <p class="text-[11px] text-gray-500 leading-snug mb-2">{{ arch.source.note }}</p>
      <PeekPanel label="YAML sections" icon="i-lucide-list" :count="arch.yamlSections.length" accent="amber">
        <ol class="space-y-1.5 pl-0 list-none">
          <li v-for="s in arch.yamlSections" :key="s.name" class="flex items-baseline gap-2">
            <code class="text-[11px] font-mono font-semibold text-amber-700 flex-shrink-0">{{ s.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ s.desc }}</span>
            <EvidenceChip v-if="s.refs" :refs="s.refs" :catalog="SOURCES" size="xs" />
          </li>
        </ol>
      </PeekPanel>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- Stage 2: Modeling — joins / dimensions / measures -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="amber" :items="arch.insights.model" />
    </div>
    <div>
      <Connector label="parse + validate" />
      <ArchBox icon="i-lucide-shapes" title="Relation + Metric Model" role="语义建模" accent="amber">
        <template #refs>
          <EvidenceChip :refs="['S2', 'S4']" :catalog="SOURCES" />
        </template>
        <div class="space-y-2">
          <div class="rounded-xl border border-amber-200 bg-amber-50/40 p-2.5">
            <div class="flex items-center gap-1.5 mb-1.5">
              <div class="i-lucide-spline text-amber-500 text-sm" />
              <span class="text-xs font-bold text-gray-700">Joins</span>
              <span class="text-[10px] text-gray-400 ml-auto">star / snowflake</span>
            </div>
            <PeekPanel label="join modes & hints" icon="i-lucide-info" :count="arch.joinModes.length" accent="amber">
              <ol class="space-y-1.5 pl-0 list-none">
                <li v-for="j in arch.joinModes" :key="j.name" class="flex items-baseline gap-2">
                  <code class="text-[11px] font-mono font-semibold text-amber-700 flex-shrink-0">{{ j.name }}</code>
                  <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ j.desc }}</span>
                  <EvidenceChip v-if="j.refs" :refs="j.refs" :catalog="SOURCES" size="xs" />
                </li>
              </ol>
            </PeekPanel>
          </div>
          <div class="rounded-xl border border-amber-200 bg-amber-50/40 p-2.5">
            <div class="flex items-center gap-1.5 mb-1.5">
              <div class="i-lucide-list-ordered text-amber-500 text-sm" />
              <span class="text-xs font-bold text-gray-700">Modeling features</span>
            </div>
            <ul class="space-y-1 pl-0 list-none">
              <li v-for="f in arch.modelingFeatures" :key="f.name" class="flex items-baseline gap-2">
                <code class="text-[11px] font-mono font-semibold text-amber-700 flex-shrink-0">{{ f.name }}</code>
                <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ f.desc }}</span>
                <EvidenceChip v-if="f.refs" :refs="f.refs" :catalog="SOURCES" size="xs" />
              </li>
            </ul>
          </div>
        </div>
      </ArchBox>
    </div>
    <!-- right: YAML demo -->
    <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
      <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
        <div class="i-lucide-file-code-2 text-amber-500 text-sm" />
        <span class="text-sm font-bold text-slate-800">YAML Definition</span>
        <span class="text-[11px] text-slate-400 ml-auto font-mono">main.sales.revenue</span>
        <EvidenceChip :refs="['S4']" :catalog="SOURCES" size="xs" />
      </div>
      <pre class="text-[10.5px] leading-relaxed font-mono text-gray-700 p-3 overflow-auto whitespace-pre">{{ arch.yamlExample }}</pre>
    </div>

    <!-- Stage 3: UC Object props -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="indigo" :items="arch.insights.object" />
    </div>
    <div>
      <Connector label="persisted as UC object" />
      <ArchBox icon="i-lucide-database" title="Metric View as UC Object" role="DDL / GRANT / lineage" accent="indigo">
        <template #refs>
          <EvidenceChip :refs="['S1', 'S2']" :catalog="SOURCES" />
        </template>
        <ul class="space-y-1.5 pl-0 list-none">
          <li v-for="p in arch.ucProps" :key="p.name" class="flex items-baseline gap-2">
            <code class="text-[11px] font-mono font-semibold text-indigo-700 flex-shrink-0">{{ p.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ p.desc }}</span>
            <EvidenceChip v-if="p.refs" :refs="p.refs" :catalog="SOURCES" size="xs" />
          </li>
        </ul>
      </ArchBox>
    </div>
    <!-- right: rewritten SQL demo -->
    <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
      <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
        <div class="i-lucide-replace text-violet-500 text-sm" />
        <span class="text-sm font-bold text-slate-800">引擎自动重写为聚合 SQL</span>
        <EvidenceChip :refs="['S2']" :catalog="SOURCES" size="xs" />
      </div>
      <pre class="text-[10.5px] leading-relaxed font-mono text-gray-700 p-3 overflow-auto whitespace-pre">{{ arch.rewrittenSql }}</pre>
    </div>
  </div>
</template>
