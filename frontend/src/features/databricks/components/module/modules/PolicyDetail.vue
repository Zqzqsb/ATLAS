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
const arch = computed(() => getDbxModule(props.flow.id)?.policy ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- Stage 1: input — base table policies -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="emerald" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-table-2" :title="arch.input.label" role="policies live here" accent="emerald">
      <template #refs>
        <EvidenceChip :refs="arch.input.refs" :catalog="SOURCES" />
      </template>
      <p class="text-[11px] text-gray-500 leading-snug mb-2">{{ arch.input.note }}</p>
      <ul class="space-y-1.5 pl-0 list-none">
        <li v-for="p in arch.baseTablePolicies" :key="p.name" class="flex items-baseline gap-2">
          <code class="text-[11px] font-mono font-semibold text-emerald-700 flex-shrink-0">{{ p.name }}</code>
          <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ p.desc }}</span>
          <EvidenceChip v-if="p.refs" :refs="p.refs" :catalog="SOURCES" size="xs" />
        </li>
      </ul>
    </ArchBox>
    <!-- right: DDL example -->
    <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
      <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
        <div class="i-lucide-shield text-emerald-500 text-sm" />
        <span class="text-sm font-bold text-slate-800">Row Filter / Column Mask DDL</span>
        <EvidenceChip class="ml-auto" :refs="['S6']" :catalog="SOURCES" size="xs" />
      </div>
      <pre class="text-[10.5px] leading-relaxed font-mono text-gray-700 p-3 overflow-auto whitespace-pre">{{ arch.ddlExample }}</pre>
    </div>

    <!-- Stage 2: propagation rule -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="emerald" :items="arch.insights.base" />
    </div>
    <div>
      <Connector label="查询 metric view 时" />
      <ArchBox icon="i-lucide-share-2" title="Runtime Propagation" role="底表策略向上传播" accent="emerald">
        <template #refs>
          <EvidenceChip :refs="arch.propagation.refs" :catalog="SOURCES" />
        </template>
        <ul class="space-y-1.5 pl-0 list-none">
          <li v-for="(p, i) in arch.propagation.points" :key="i" class="flex items-start gap-2">
            <div class="i-lucide-arrow-up text-emerald-400 text-xs mt-1 flex-shrink-0" />
            <span class="text-[11px] text-gray-600 leading-snug">{{ p }}</span>
          </li>
        </ul>
        <div class="mt-2 rounded-lg border border-emerald-100 bg-emerald-50/40 px-2.5 py-1.5">
          <div class="flex items-center gap-1.5 mb-1">
            <div class="i-lucide-key-round text-emerald-500 text-xs" />
            <span class="text-[11px] font-bold text-emerald-700">Object ACL（独立体系）</span>
          </div>
          <ul class="space-y-1 pl-0 list-none">
            <li v-for="o in arch.objectAcl" :key="o.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-emerald-700 flex-shrink-0">{{ o.name }}</code>
              <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ o.desc }}</span>
              <EvidenceChip v-if="o.refs" :refs="o.refs" :catalog="SOURCES" size="xs" />
            </li>
          </ul>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />

    <!-- Stage 3: NOT policy-aware (D-graded) -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="emerald" :items="arch.insights.propagate" />
    </div>
    <div>
      <Connector label="generation phase" />
      <ArchBox icon="i-lucide-eye-off" title="NOT policy-aware @ generation" role="生成阶段不感知策略" accent="slate" muted>
        <template #refs>
          <EvidenceChip v-if="arch.notPolicyAware.refs" :refs="arch.notPolicyAware.refs" :catalog="SOURCES" />
        </template>
        <ul class="space-y-1.5 pl-0 list-none">
          <li v-for="(p, i) in arch.notPolicyAware.points" :key="i" class="flex items-start gap-2">
            <div class="i-lucide-x text-gray-400 text-xs mt-1 flex-shrink-0" />
            <span class="text-[11px] text-gray-500 leading-snug">{{ p }}</span>
          </li>
        </ul>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
