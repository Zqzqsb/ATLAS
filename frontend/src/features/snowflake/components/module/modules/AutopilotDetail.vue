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
const arch = computed(() => getSnowModule(props.flow.id)?.autopilot ?? null)

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
    <ArchBox icon="i-lucide-history" :title="arch.input.label" role="generation source" accent="violet">
      <template #refs>
        <EvidenceChip :refs="arch.input.refs" :catalog="SOURCES" />
      </template>
      <p class="text-[11px] text-gray-500 leading-snug">{{ arch.input.note }}</p>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- Stage 2: three AI assistance paths -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="violet" :items="arch.insights.generate" />
    </div>
    <div>
      <Connector label="three AI paths" />
      <ArchBox icon="i-lucide-sparkles" title="AI Assistance" role="生成 + 建议 + 工具" accent="violet">
        <template #refs>
          <EvidenceChip :refs="['S11', 'S12', 'S13']" :catalog="SOURCES" />
        </template>
        <div class="space-y-2">
          <div class="rounded-lg border border-violet-200 bg-violet-50/40 px-2.5 py-1.5">
            <div class="flex items-center gap-1.5 mb-1">
              <div class="i-lucide-rocket text-violet-500 text-sm" />
              <span class="text-xs font-bold text-violet-700">Semantic View Autopilot</span>
              <EvidenceChip class="ml-auto" :refs="['S11']" :catalog="SOURCES" size="xs" />
            </div>
            <ul class="space-y-1 pl-0 list-none">
              <li v-for="a in arch.autopilot" :key="a.name" class="flex items-baseline gap-2">
                <code class="text-[11px] font-mono font-semibold text-violet-700 flex-shrink-0">{{ a.name }}</code>
                <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ a.desc }}</span>
                <EvidenceChip v-if="a.refs" :refs="a.refs" :catalog="SOURCES" size="xs" />
              </li>
            </ul>
          </div>
          <div class="rounded-lg border border-violet-200 bg-violet-50/40 px-2.5 py-1.5">
            <div class="flex items-center gap-1.5 mb-1">
              <div class="i-lucide-list-checks text-violet-500 text-sm" />
              <span class="text-xs font-bold text-violet-700">Snowsight Suggestions</span>
              <EvidenceChip class="ml-auto" :refs="['S12']" :catalog="SOURCES" size="xs" />
            </div>
            <ul class="space-y-1 pl-0 list-none">
              <li v-for="s in arch.suggestions" :key="s.name" class="flex items-baseline gap-2">
                <code class="text-[11px] font-mono font-semibold text-violet-700 flex-shrink-0">{{ s.name }}</code>
                <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ s.desc }}</span>
                <EvidenceChip v-if="s.refs" :refs="s.refs" :catalog="SOURCES" size="xs" />
              </li>
            </ul>
          </div>
          <div class="rounded-lg border border-violet-200 bg-violet-50/40 px-2.5 py-1.5">
            <div class="flex items-center gap-1.5 mb-1">
              <div class="i-lucide-package text-violet-500 text-sm" />
              <span class="text-xs font-bold text-violet-700">semantic-model-generator (Labs)</span>
              <EvidenceChip class="ml-auto" :refs="['S13']" :catalog="SOURCES" size="xs" />
            </div>
            <ul class="space-y-1 pl-0 list-none">
              <li v-for="g in arch.generator" :key="g.name" class="flex items-baseline gap-2">
                <code class="text-[11px] font-mono font-semibold text-violet-700 flex-shrink-0">{{ g.name }}</code>
                <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ g.desc }}</span>
                <EvidenceChip v-if="g.refs" :refs="g.refs" :catalog="SOURCES" size="xs" />
              </li>
            </ul>
          </div>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />

    <!-- Stage 3: human review gate -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="emerald" :items="arch.insights.review" />
    </div>
    <div>
      <Connector label="human review queue" />
      <ArchBox icon="i-lucide-user-check" title="Review Gate" role="人工接受 / 编辑 / 驳回" accent="emerald">
        <template #refs>
          <EvidenceChip :refs="arch.reviewGate.refs" :catalog="SOURCES" />
        </template>
        <ul class="space-y-1.5 pl-0 list-none">
          <li v-for="(p, i) in arch.reviewGate.points" :key="i" class="flex items-start gap-2">
            <div class="i-lucide-user text-emerald-400 text-xs mt-1 flex-shrink-0" />
            <span class="text-[11px] text-gray-600 leading-snug">{{ p }}</span>
          </li>
        </ul>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
