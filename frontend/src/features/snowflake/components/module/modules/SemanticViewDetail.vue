<script setup lang="ts">
import { computed } from 'vue'
import ArchBox from '../../../../arch/components/module/diagram/ArchBox.vue'
import Connector from '../../../../arch/components/module/diagram/Connector.vue'
import PeekPanel from '../../../../arch/components/module/diagram/PeekPanel.vue'
import InsightNotes from '../../../../arch/components/module/diagram/InsightNotes.vue'
import EvidenceChip from '../../../../arch/components/module/diagram/EvidenceChip.vue'
import { SOURCES } from '../../../model/sources'
import type { SnowFlowDef } from '../../../model/flows'
import { getSnowModule } from '../../../model/modules'

const props = defineProps<{ flow: SnowFlowDef; showNotes?: boolean }>()
const arch = computed(() => getSnowModule(props.flow.id)?.semanticView ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- Stage 1: forms -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="amber" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-shapes" title="Two Forms" role="DDL or YAML" accent="amber">
      <template #refs>
        <EvidenceChip :refs="['S1', 'S2']" :catalog="SOURCES" />
      </template>
      <ul class="space-y-1.5 pl-0 list-none">
        <li v-for="f in arch.forms" :key="f.name" class="rounded-lg border px-2.5 py-1.5" :class="f.recommended ? 'border-amber-300 bg-amber-50/60' : 'border-gray-200 bg-gray-50/40'">
          <div class="flex items-center gap-2 mb-0.5">
            <code class="text-[11px] font-mono font-bold" :class="f.recommended ? 'text-amber-700' : 'text-gray-600'">{{ f.name }}</code>
            <span v-if="f.recommended" class="px-1 py-0.5 rounded text-[9px] font-bold bg-amber-100 text-amber-700">推荐</span>
            <EvidenceChip v-if="f.refs" class="ml-auto" :refs="f.refs" :catalog="SOURCES" size="xs" />
          </div>
          <span class="text-[11px] text-gray-500 leading-snug">{{ f.desc }}</span>
        </li>
      </ul>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- Stage 2: DDL sections + joins + metrics -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="amber" :items="arch.insights.model" />
    </div>
    <div>
      <Connector label="CREATE SEMANTIC VIEW" />
      <ArchBox icon="i-lucide-list-ordered" title="Semantic View Structure" role="关系 + 指标模型" accent="amber">
        <template #refs>
          <EvidenceChip :refs="['S3', 'S4']" :catalog="SOURCES" />
        </template>
        <PeekPanel label="DDL sections" icon="i-lucide-list" :count="arch.ddlSections.length" accent="amber">
          <ol class="space-y-1.5 pl-0 list-none">
            <li v-for="s in arch.ddlSections" :key="s.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-amber-700 flex-shrink-0">{{ s.name }}</code>
              <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ s.desc }}</span>
              <EvidenceChip v-if="s.refs" :refs="s.refs" :catalog="SOURCES" size="xs" />
            </li>
          </ol>
        </PeekPanel>
        <div class="grid grid-cols-2 gap-2 mt-2">
          <PeekPanel label="join support" icon="i-lucide-spline" :count="arch.joinSupport.length" accent="blue">
            <ul class="space-y-1 pl-0 list-none">
              <li v-for="j in arch.joinSupport" :key="j.name" class="flex items-baseline gap-2">
                <code class="text-[11px] font-mono font-semibold text-blue-700 flex-shrink-0">{{ j.name }}</code>
                <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ j.desc }}</span>
                <EvidenceChip v-if="j.refs" :refs="j.refs" :catalog="SOURCES" size="xs" />
              </li>
            </ul>
          </PeekPanel>
          <PeekPanel label="metric features" icon="i-lucide-calculator" :count="arch.metricFeatures.length" accent="violet">
            <ul class="space-y-1 pl-0 list-none">
              <li v-for="m in arch.metricFeatures" :key="m.name" class="flex items-baseline gap-2">
                <code class="text-[11px] font-mono font-semibold text-violet-700 flex-shrink-0">{{ m.name }}</code>
                <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ m.desc }}</span>
                <EvidenceChip v-if="m.refs" :refs="m.refs" :catalog="SOURCES" size="xs" />
              </li>
            </ul>
          </PeekPanel>
        </div>
      </ArchBox>
    </div>
    <!-- right: DDL + YAML demo -->
    <div class="space-y-3">
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-file-code-2 text-amber-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">DDL · CREATE SEMANTIC VIEW</span>
          <EvidenceChip class="ml-auto" :refs="['S4']" :catalog="SOURCES" size="xs" />
        </div>
        <pre class="text-[10.5px] leading-relaxed font-mono text-gray-700 p-3 overflow-auto whitespace-pre">{{ arch.ddlExample }}</pre>
      </div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-file-text text-amber-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">YAML on stage（兼容形态）</span>
          <EvidenceChip class="ml-auto" :refs="['S17']" :catalog="SOURCES" size="xs" />
        </div>
        <pre class="text-[10.5px] leading-relaxed font-mono text-gray-700 p-3 overflow-auto whitespace-pre">{{ arch.yamlExample }}</pre>
      </div>
    </div>

    <!-- Stage 3: governance -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="emerald" :items="arch.insights.form" />
    </div>
    <div>
      <Connector label="treated as schema-level metadata" />
      <ArchBox icon="i-lucide-shield-check" title="Governance" role="DDL / GRANT / RBAC" accent="emerald">
        <template #refs>
          <EvidenceChip :refs="['S2', 'S15']" :catalog="SOURCES" />
        </template>
        <ul class="space-y-1.5 pl-0 list-none">
          <li v-for="g in arch.governance" :key="g.name" class="flex items-baseline gap-2">
            <code class="text-[11px] font-mono font-semibold text-emerald-700 flex-shrink-0">{{ g.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ g.desc }}</span>
            <EvidenceChip v-if="g.refs" :refs="g.refs" :catalog="SOURCES" size="xs" />
          </li>
        </ul>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
