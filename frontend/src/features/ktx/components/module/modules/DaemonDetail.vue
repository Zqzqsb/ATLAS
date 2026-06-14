<script setup lang="ts">
import { computed } from 'vue'
import { ACCENTS } from '../../../../arch/model/architecture'
import ArchBox from '../../../../arch/components/module/diagram/ArchBox.vue'
import Connector from '../../../../arch/components/module/diagram/Connector.vue'
import InsightNotes from '../../../../arch/components/module/diagram/InsightNotes.vue'
import type { KtxFlowDef } from '../../../model/ktx'
import { getKtxModule } from '../../../model/ktx'

const props = defineProps<{ flow: KtxFlowDef; showNotes?: boolean }>()
const arch = computed(() => getKtxModule(props.flow.id)?.daemon ?? null)

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
    <ArchBox icon="i-lucide-cog" :title="arch.input.label" accent="slate" muted>
      <div class="text-[11px] text-gray-500 leading-snug">{{ arch.input.note }}</div>
    </ArchBox>
    <div class="hidden lg:block" />

    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="rose" :items="arch.insights.bridge" />
    </div>
    <div>
      <Connector label="FastAPI routes" />
      <ArchBox icon="i-lucide-server" title="ktx-daemon Endpoints" accent="rose" badge="× 4">
        <div class="space-y-2">
          <div
            v-for="g in arch.endpoints"
            :key="g.group"
            class="rounded-xl border p-2.5"
            :class="ACCENTS[g.accent].surface"
          >
            <div class="flex items-center gap-1.5 mb-1.5">
              <div :class="[g.icon, ACCENTS[g.accent].text, 'text-sm']" />
              <span class="text-xs font-bold text-gray-700">{{ g.group }}</span>
            </div>
            <div v-for="r in g.routes" :key="r.path" class="flex items-baseline gap-2 mb-0.5">
              <code class="text-[10px] font-mono font-semibold flex-shrink-0" :class="ACCENTS[g.accent].text">{{ r.path }}</code>
              <span class="text-[10px] text-gray-500">{{ r.desc }}</span>
            </div>
          </div>
        </div>
      </ArchBox>
    </div>
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-terminal text-rose-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">启动方式</span>
        </div>
        <pre class="text-[10px] leading-relaxed font-mono text-gray-600 p-3 whitespace-pre-wrap">{{ arch.startup }}</pre>
        <div class="border-t border-slate-100 p-2.5 space-y-1">
          <div v-for="p in arch.pools" :key="p.name" class="flex items-center gap-2 px-2 py-1 rounded-md bg-rose-50/30">
            <div :class="[p.icon, 'text-rose-500 text-xs']" />
            <code class="text-[10.5px] font-mono font-bold text-gray-700">{{ p.name }}</code>
            <span class="text-[10px] text-gray-500 ml-auto">{{ p.desc }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="slate" :items="arch.insights.isolate" />
    </div>
    <div>
      <Connector label="TS ↔ Python HTTP" />
      <ArchBox icon="i-lucide-bridge" title="端口映射" accent="slate">
        <div class="space-y-2">
          <div v-for="p in arch.ports" :key="p.tsPort" class="rounded-lg border border-slate-200 bg-slate-50/50 px-2.5 py-2">
            <code class="text-[11px] font-mono font-bold text-gray-800">{{ p.tsPort }}</code>
            <div class="text-[10.5px] text-gray-500">
              <span class="font-mono text-rose-600">{{ p.httpRoute }}</span> → {{ p.usedBy }}
            </div>
          </div>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
