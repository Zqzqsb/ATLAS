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
const arch = computed(() => getDbxModule(props.flow.id)?.genie ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- Stage 1: Curation (Space) -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="slate" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-sparkles" :title="arch.input.label" role="curated workspace" accent="slate">
      <template #refs>
        <EvidenceChip :refs="arch.input.refs" :catalog="SOURCES" />
      </template>
      <p class="text-[11px] text-gray-500 leading-snug mb-2">{{ arch.input.note }}</p>
      <div class="space-y-1.5">
        <div v-for="c in arch.curation" :key="c.name" class="rounded-lg border border-slate-200 bg-slate-50/60 px-2.5 py-1.5">
          <div class="flex items-center gap-2 mb-0.5">
            <code class="text-[11px] font-mono font-bold text-slate-700">{{ c.name }}</code>
            <EvidenceChip v-if="c.refs" class="ml-auto" :refs="c.refs" :catalog="SOURCES" size="xs" />
          </div>
          <span class="text-[11px] text-gray-500 leading-snug">{{ c.desc }}</span>
        </div>
      </div>
    </ArchBox>
    <div v-if="showNotes" class="hidden lg:block" />
    <div v-else class="hidden lg:block" />

    <!-- Stage 2: Genie pipeline (route → generate → execute → return) -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="slate" :items="arch.insights.runtime" />
    </div>
    <div>
      <Connector label="ask Genie / Conversation API" />
      <ArchBox icon="i-lucide-workflow" title="Genie Pipeline" role="NL → SQL → run" accent="slate">
        <template #refs>
          <EvidenceChip :refs="['S7', 'S8', 'S10', 'S13']" :catalog="SOURCES" />
        </template>
        <ol class="space-y-1.5 pl-0 list-none">
          <li v-for="(step, i) in arch.flow" :key="step.name" class="flex items-baseline gap-2">
            <span class="text-[10px] font-mono font-bold text-slate-400 flex-shrink-0">{{ i + 1 }}.</span>
            <code class="text-[11px] font-mono font-semibold text-slate-700 flex-shrink-0">{{ step.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ step.desc }}</span>
            <EvidenceChip v-if="step.refs" :refs="step.refs" :catalog="SOURCES" size="xs" />
          </li>
        </ol>
        <div class="grid grid-cols-2 gap-2 mt-2.5">
          <PeekPanel label="guards" icon="i-lucide-shield" :count="arch.guards.length" accent="emerald">
            <ul class="space-y-1 pl-0 list-none">
              <li v-for="g in arch.guards" :key="g.name" class="flex items-baseline gap-2">
                <code class="text-[11px] font-mono font-semibold text-emerald-700 flex-shrink-0">{{ g.name }}</code>
                <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ g.desc }}</span>
                <EvidenceChip v-if="g.refs" :refs="g.refs" :catalog="SOURCES" size="xs" />
              </li>
            </ul>
          </PeekPanel>
          <PeekPanel label="explain" icon="i-lucide-eye" :count="arch.explain.length" accent="violet">
            <ul class="space-y-1 pl-0 list-none">
              <li v-for="e in arch.explain" :key="e.name" class="flex items-baseline gap-2">
                <code class="text-[11px] font-mono font-semibold text-violet-700 flex-shrink-0">{{ e.name }}</code>
                <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ e.desc }}</span>
                <EvidenceChip v-if="e.refs" :refs="e.refs" :catalog="SOURCES" size="xs" />
              </li>
            </ul>
          </PeekPanel>
        </div>
      </ArchBox>
    </div>
    <!-- right: sample conversation -->
    <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
      <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
        <div class="i-lucide-message-square text-slate-500 text-sm" />
        <span class="text-sm font-bold text-slate-800">示例对话（mock）</span>
        <span v-if="arch.sampleConvo.verified" class="ml-auto px-1.5 py-0.5 rounded text-[9px] font-mono font-bold bg-emerald-50 text-emerald-700 border border-emerald-200">verified</span>
      </div>
      <div class="px-3 py-2 border-b border-slate-100">
        <div class="text-[10px] text-slate-400 mb-0.5">USER</div>
        <div class="text-xs font-semibold text-slate-700 leading-snug">{{ arch.sampleConvo.user }}</div>
      </div>
      <div class="px-3 py-2 border-b border-slate-100 bg-slate-50/60">
        <div class="text-[10px] text-slate-400 mb-0.5">INTERPRETATION</div>
        <div class="text-[11px] text-slate-600 leading-snug italic">{{ arch.sampleConvo.interp }}</div>
      </div>
      <pre class="text-[10.5px] leading-relaxed font-mono text-gray-700 p-3 overflow-auto whitespace-pre">{{ arch.sampleConvo.sql }}</pre>
    </div>

    <!-- Stage 3: Feedback / verified -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="emerald" :items="arch.insights.feedback" />
    </div>
    <div>
      <Connector label="user feedback / manager review" />
      <ArchBox icon="i-lucide-recycle" title="Feedback Loop" role="trusted asset" accent="emerald">
        <template #refs>
          <EvidenceChip :refs="['S8', 'S9']" :catalog="SOURCES" />
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
    <!-- right: undisclosed -->
    <div class="rounded-2xl border border-dashed border-gray-300 bg-white px-3.5 py-3">
      <div class="flex items-center gap-1.5 mb-2">
        <div class="i-lucide-eye-off text-gray-400 text-sm" />
        <span class="text-xs font-bold text-gray-700">内部未披露 · 公开文档无证据</span>
        <span class="ml-auto px-1.5 py-0.5 rounded text-[9px] font-mono font-bold bg-gray-100 text-gray-500 border border-gray-200">D</span>
      </div>
      <ul class="space-y-1.5 pl-0 list-none">
        <li v-for="u in arch.undisclosed" :key="u.name" class="flex flex-col gap-0.5">
          <code class="text-[11px] font-mono font-semibold text-gray-500">{{ u.name }}</code>
          <span class="text-[11px] text-gray-400 leading-snug">{{ u.desc }}</span>
        </li>
      </ul>
    </div>
  </div>
</template>
