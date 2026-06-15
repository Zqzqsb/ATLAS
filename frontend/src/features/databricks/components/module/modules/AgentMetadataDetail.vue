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
const arch = computed(() => getDbxModule(props.flow.id)?.agentMetadata ?? null)

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
      <InsightNotes accent="amber" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-tag" :title="arch.input.label" role="UC objects" accent="amber" muted>
      <template #refs>
        <EvidenceChip :refs="arch.input.refs" :catalog="SOURCES" />
      </template>
      <p class="text-[11px] text-gray-500 leading-snug">{{ arch.input.note }}</p>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- Stage 2: fields -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="amber" :items="arch.insights.field" />
    </div>
    <div>
      <Connector label="annotate columns" />
      <ArchBox icon="i-lucide-tags" title="Agent Metadata Fields" role="列级注解" accent="amber" :badge="`× ${arch.fields.length}`">
        <template #refs>
          <EvidenceChip :refs="['S3']" :catalog="SOURCES" />
        </template>
        <ul class="space-y-1.5 pl-0 list-none">
          <li v-for="f in arch.fields" :key="f.name" class="flex flex-col gap-0.5 rounded-lg border border-amber-100 bg-amber-50/40 px-2 py-1.5">
            <div class="flex items-center gap-2">
              <code class="text-[11px] font-mono font-bold text-amber-700">{{ f.name }}</code>
              <code v-if="f.example" class="ml-auto text-[10px] font-mono text-amber-600 truncate">{{ f.example }}</code>
              <EvidenceChip v-if="f.refs" :refs="f.refs" :catalog="SOURCES" size="xs" />
            </div>
            <span class="text-[11px] text-gray-500 leading-snug">{{ f.desc }}</span>
          </li>
        </ul>
      </ArchBox>
    </div>
    <!-- right: yaml demo -->
    <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
      <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
        <div class="i-lucide-file-text text-amber-500 text-sm" />
        <span class="text-sm font-bold text-slate-800">YAML 内联示例</span>
        <span class="text-[11px] text-slate-400 ml-auto font-mono">dimensions[]</span>
        <EvidenceChip :refs="['S3', 'S4']" :catalog="SOURCES" size="xs" />
      </div>
      <pre class="text-[10.5px] leading-relaxed font-mono text-gray-700 p-3 overflow-auto whitespace-pre">{{ arch.yamlExample }}</pre>
    </div>

    <!-- Stage 3: consumers -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="slate" :items="arch.insights.consume" />
    </div>
    <div>
      <Connector label="consumed by" />
      <ArchBox icon="i-lucide-share-2" title="Consumers" role="读取者" accent="slate">
        <template #refs>
          <EvidenceChip :refs="['S3', 'S7']" :catalog="SOURCES" />
        </template>
        <ul class="space-y-1.5 pl-0 list-none">
          <li v-for="c in arch.consumers" :key="c.name" class="flex items-baseline gap-2">
            <code class="text-[11px] font-mono font-semibold text-slate-700 flex-shrink-0">{{ c.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug flex-1">{{ c.desc }}</span>
            <EvidenceChip v-if="c.refs" :refs="c.refs" :catalog="SOURCES" size="xs" />
          </li>
        </ul>
        <div class="mt-2.5 flex items-start gap-2 rounded-lg border border-dashed border-amber-200 bg-amber-50/40 px-2.5 py-1.5">
          <div class="i-lucide-info text-amber-400 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-amber-700 leading-snug">{{ arch.retrievalNote }}</span>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
