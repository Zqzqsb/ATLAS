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
const arch = computed(() => getSnowModule(props.flow.id)?.analystFlow ?? null)

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
      <InsightNotes accent="slate" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-radio" :title="arch.input.label" role="REST API request" accent="slate">
      <template #refs>
        <EvidenceChip :refs="arch.input.refs" :catalog="SOURCES" />
      </template>
      <p class="text-[11px] text-gray-500 leading-snug">{{ arch.input.note }}</p>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- Stage 2: 6-step pipeline -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="slate" :items="arch.insights.pipeline" />
    </div>
    <div>
      <Connector label="cortex-analyst pipeline" />
      <ArchBox icon="i-lucide-workflow" title="6-Step Pipeline" role="logical → physical → check" accent="slate">
        <template #refs>
          <EvidenceChip :refs="['S9']" :catalog="SOURCES" />
        </template>
        <ol class="space-y-1.5 pl-0 list-none">
          <li v-for="step in arch.pipeline" :key="step.name" class="flex items-baseline gap-2">
            <code class="text-[11px] font-mono font-semibold text-slate-700 flex-shrink-0">{{ step.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ step.desc }}</span>
            <EvidenceChip v-if="step.refs" :refs="step.refs" :catalog="SOURCES" size="xs" />
          </li>
        </ol>
        <div class="grid grid-cols-2 gap-2 mt-2.5">
          <PeekPanel label="error correction" icon="i-lucide-bug-off" :count="arch.errorLoop.length" accent="emerald">
            <ul class="space-y-1 pl-0 list-none">
              <li v-for="e in arch.errorLoop" :key="e.name" class="flex items-baseline gap-2">
                <code class="text-[11px] font-mono font-semibold text-emerald-700 flex-shrink-0">{{ e.name }}</code>
                <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ e.desc }}</span>
                <EvidenceChip v-if="e.refs" :refs="e.refs" :catalog="SOURCES" size="xs" />
              </li>
            </ul>
          </PeekPanel>
          <PeekPanel label="response types" icon="i-lucide-file-output" :count="arch.responseTypes.length" accent="violet">
            <ul class="space-y-1 pl-0 list-none">
              <li v-for="r in arch.responseTypes" :key="r.name" class="flex items-baseline gap-2">
                <code class="text-[11px] font-mono font-semibold text-violet-700 flex-shrink-0">{{ r.name }}</code>
                <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ r.desc }}</span>
                <EvidenceChip v-if="r.refs" :refs="r.refs" :catalog="SOURCES" size="xs" />
              </li>
            </ul>
          </PeekPanel>
        </div>
      </ArchBox>
    </div>
    <!-- right: API example -->
    <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
      <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
        <div class="i-lucide-code-2 text-slate-500 text-sm" />
        <span class="text-sm font-bold text-slate-800">REST API · 请求 / 响应</span>
        <EvidenceChip class="ml-auto" :refs="['S8']" :catalog="SOURCES" size="xs" />
      </div>
      <pre class="text-[10.5px] leading-relaxed font-mono text-gray-700 p-3 overflow-auto whitespace-pre">{{ arch.apiExample }}</pre>
    </div>

    <!-- Stage 3: classification + suggestions (clarification) -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="violet" :items="arch.insights.verify" />
    </div>
    <div>
      <Connector label="when ambiguous" />
      <ArchBox icon="i-lucide-message-circle-question" title="Classification + Suggestions" role="拒答 + 相似问题" accent="violet">
        <template #refs>
          <EvidenceChip :refs="arch.suggestions.refs" :catalog="SOURCES" />
        </template>
        <ul class="space-y-1.5 pl-0 list-none">
          <li v-for="(p, i) in arch.suggestions.points" :key="i" class="flex items-start gap-2">
            <div class="i-lucide-circle-dot text-violet-400 text-xs mt-1 flex-shrink-0" />
            <span class="text-[11px] text-gray-600 leading-snug">{{ p }}</span>
          </li>
        </ul>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />

    <!-- Stage 4: safety boundary (API not execute) -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="emerald" :items="arch.insights.boundary" />
    </div>
    <div>
      <Connector label="return SQL · caller executes" />
      <ArchBox icon="i-lucide-shield" title="Safety Boundary" role="API 不执行" accent="emerald">
        <template #refs>
          <EvidenceChip :refs="['S1', 'S8']" :catalog="SOURCES" />
        </template>
        <ul class="space-y-1.5 pl-0 list-none">
          <li class="flex items-start gap-2">
            <div class="i-lucide-arrow-right text-emerald-400 text-xs mt-1" />
            <span class="text-[11px] text-gray-600 leading-snug">REST API 仅生成并返回 SQL · 调用方在自身 warehouse 用自身角色执行</span>
          </li>
          <li class="flex items-start gap-2">
            <div class="i-lucide-arrow-right text-emerald-400 text-xs mt-1" />
            <span class="text-[11px] text-gray-600 leading-snug">默认 Snowflake 托管 LLM · 数据 / 元数据 / 提示不离开 governance boundary</span>
          </li>
          <li class="flex items-start gap-2">
            <div class="i-lucide-arrow-right text-emerald-400 text-xs mt-1" />
            <span class="text-[11px] text-gray-600 leading-snug">承诺不在客户数据上训练（runtime / 平台级安全边界）</span>
          </li>
        </ul>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
