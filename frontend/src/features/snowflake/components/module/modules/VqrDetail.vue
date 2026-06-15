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
const arch = computed(() => getSnowModule(props.flow.id)?.vqr ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- Stage 1: VQR + fields -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="violet" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-shield-check" :title="arch.input.label" role="人工审核过的 NL-SQL 对" accent="violet">
      <template #refs>
        <EvidenceChip :refs="arch.input.refs" :catalog="SOURCES" />
      </template>
      <p class="text-[11px] text-gray-500 leading-snug mb-2">{{ arch.input.note }}</p>
      <ul class="space-y-1 pl-0 list-none">
        <li v-for="f in arch.fields" :key="f.name" class="flex items-baseline gap-2">
          <code class="text-[11px] font-mono font-bold text-violet-700 flex-shrink-0">{{ f.name }}</code>
          <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ f.desc }}</span>
          <EvidenceChip v-if="f.refs" :refs="f.refs" :catalog="SOURCES" size="xs" />
        </li>
      </ul>
    </ArchBox>
    <!-- right: yaml example -->
    <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
      <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
        <div class="i-lucide-file-text text-violet-500 text-sm" />
        <span class="text-sm font-bold text-slate-800">verified_queries · custom_instructions</span>
        <EvidenceChip class="ml-auto" :refs="['S5', 'S14']" :catalog="SOURCES" size="xs" />
      </div>
      <pre class="text-[10.5px] leading-relaxed font-mono text-gray-700 p-3 overflow-auto whitespace-pre">{{ arch.yamlExample }}</pre>
    </div>

    <!-- Stage 2: add paths + custom instructions -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="violet" :items="arch.insights.repo" />
    </div>
    <div>
      <Connector label="three add paths" />
      <ArchBox icon="i-lucide-plus" title="Add Verified Queries" role="三种添加路径" accent="violet">
        <template #refs>
          <EvidenceChip :refs="['S5', 'S12']" :catalog="SOURCES" />
        </template>
        <ol class="space-y-1.5 pl-0 list-none">
          <li v-for="(p, i) in arch.addPaths" :key="p.name" class="flex items-baseline gap-2">
            <span class="text-[10px] font-mono font-bold text-violet-400 flex-shrink-0">{{ i + 1 }}.</span>
            <code class="text-[11px] font-mono font-semibold text-violet-700 flex-shrink-0">{{ p.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ p.desc }}</span>
            <EvidenceChip v-if="p.refs" :refs="p.refs" :catalog="SOURCES" size="xs" />
          </li>
        </ol>
        <div class="mt-2.5 rounded-lg border border-amber-200 bg-amber-50/40 px-2.5 py-1.5">
          <div class="flex items-center gap-1.5 mb-1">
            <div class="i-lucide-pencil text-amber-500 text-xs" />
            <span class="text-[11px] font-bold text-amber-700">Custom Instructions（并列富上下文）</span>
          </div>
          <ul class="space-y-1 pl-0 list-none">
            <li v-for="c in arch.customInstructions" :key="c.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-amber-700 flex-shrink-0">{{ c.name }}</code>
              <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ c.desc }}</span>
              <EvidenceChip v-if="c.refs" :refs="c.refs" :catalog="SOURCES" size="xs" />
            </li>
          </ul>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />

    <!-- Stage 3: feedback loop -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="emerald" :items="arch.insights.feedback" />
    </div>
    <div>
      <Connector label="user/admin feedback" />
      <ArchBox icon="i-lucide-recycle" title="Feedback Loop" role="人工驱动闭环" accent="emerald">
        <template #refs>
          <EvidenceChip :refs="['S5', 'S8', 'S16']" :catalog="SOURCES" />
        </template>
        <ul class="space-y-1.5 pl-0 list-none">
          <li v-for="fb in arch.feedback" :key="fb.name" class="flex items-baseline gap-2">
            <code class="text-[11px] font-mono font-semibold text-emerald-700 flex-shrink-0">{{ fb.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ fb.desc }}</span>
            <EvidenceChip v-if="fb.refs" :refs="fb.refs" :catalog="SOURCES" size="xs" />
          </li>
        </ul>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
