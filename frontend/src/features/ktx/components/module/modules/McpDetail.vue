<script setup lang="ts">
import { computed } from 'vue'
import { ACCENTS } from '../../../../arch/model/architecture'
import ArchBox from '../../../../arch/components/module/diagram/ArchBox.vue'
import Connector from '../../../../arch/components/module/diagram/Connector.vue'
import PeekPanel from '../../../../arch/components/module/diagram/PeekPanel.vue'
import InsightNotes from '../../../../arch/components/module/diagram/InsightNotes.vue'
import type { KtxFlowDef } from '../../../model/ktx'
import { getKtxModule } from '../../../model/ktx'

const props = defineProps<{ flow: KtxFlowDef; showNotes?: boolean }>()
const arch = computed(() => getKtxModule(props.flow.id)?.mcp ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="slate" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-bot" :title="arch.input.label" accent="slate" muted>
      <div class="flex flex-wrap gap-1 mb-1.5">
        <span
          v-for="c in arch.input.clients"
          :key="c"
          class="px-2 py-0.5 rounded-full text-[10px] font-semibold bg-slate-100 text-slate-600 border border-slate-200"
        >{{ c }}</span>
      </div>
      <div class="text-[11px] text-gray-500">{{ arch.input.note }}</div>
    </ArchBox>
    <div class="hidden lg:block" />

    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="violet" :items="arch.insights.surface" />
    </div>
    <div>
      <Connector label="ktx mcp start" />
      <ArchBox icon="i-lucide-server" title="MCP Server" role="stdio + HTTP" accent="violet">
        <div class="space-y-1.5 mb-2">
          <div v-for="t in arch.transports" :key="t.name" class="rounded-lg border border-violet-100 bg-violet-50/40 px-2.5 py-2">
            <div class="flex items-center gap-1.5">
              <div :class="[t.icon, 'text-violet-500 text-sm']" />
              <span class="text-xs font-bold text-gray-700">{{ t.name }}</span>
            </div>
            <code class="text-[10.5px] font-mono text-violet-700">{{ t.cmd }}</code>
            <div class="text-[10.5px] text-gray-500">{{ t.note }}</div>
          </div>
        </div>
        <PeekPanel label="MCP Tools" icon="i-lucide-wrench" :count="arch.tools.length" accent="violet">
          <div class="space-y-1.5">
            <div
              v-for="t in arch.tools"
              :key="t.name"
              class="rounded-lg border px-2 py-1.5"
              :class="ACCENTS[t.accent].surface"
            >
              <div class="flex items-center gap-2">
                <code class="text-[11px] font-mono font-bold" :class="ACCENTS[t.accent].text">{{ t.name }}</code>
                <span v-if="t.writes" class="text-[9px] font-bold text-rose-600 ml-auto">WRITE</span>
              </div>
              <div class="text-[10.5px] text-gray-500">{{ t.desc }}</div>
            </div>
          </div>
        </PeekPanel>
      </ArchBox>
    </div>
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-code-2 text-amber-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">{{ arch.sampleCall.tool }}</span>
        </div>
        <pre class="text-[10px] leading-relaxed font-mono text-gray-600 p-3 overflow-auto whitespace-pre">{{ arch.sampleCall.req }}</pre>
        <div class="flex items-center justify-center py-0.5 border-t border-slate-100">
          <div class="i-lucide-chevron-down text-amber-300 text-sm" />
        </div>
        <pre class="text-[10px] leading-relaxed font-mono text-gray-500 bg-amber-50/40 p-3 overflow-auto whitespace-pre">{{ arch.sampleCall.resp }}</pre>
      </div>
    </div>

    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="rose" :items="arch.insights.safety" />
    </div>
    <div>
      <Connector label="local-project-ports" />
      <ArchBox icon="i-lucide-plug" title="端口接线" accent="slate">
        <div class="space-y-1.5">
          <div v-for="p in arch.ports" :key="p.name" class="flex items-baseline gap-2">
            <code class="text-[11px] font-mono font-semibold text-gray-700 flex-shrink-0">{{ p.name }}</code>
            <span class="text-[11px] text-gray-500">{{ p.desc }}</span>
          </div>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
