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
const arch = computed(() => getKtxModule(props.flow.id)?.ingest ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="emerald" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-terminal" :title="arch.input.label" accent="slate" muted>
      <div class="text-xs text-gray-500 leading-snug">{{ arch.input.note }}</div>
    </ArchBox>
    <div class="hidden lg:block" />

    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="emerald" :items="arch.insights.pipeline" />
    </div>
    <div>
      <Connector label="8 SourceAdapters" />
      <ArchBox icon="i-lucide-plug" title="SourceAdapter Registry" role="协议化适配" accent="emerald" badge="× 8">
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-1.5">
          <div
            v-for="a in arch.adapters"
            :key="a.name"
            class="rounded-xl border p-2 flex flex-col gap-0.5"
            :class="ACCENTS[a.accent].surface"
          >
            <div class="flex items-center gap-1.5">
              <div :class="[a.icon, ACCENTS[a.accent].text, 'text-sm flex-shrink-0']" />
              <code class="text-[11px] font-mono font-bold text-gray-800">{{ a.name }}</code>
            </div>
            <span class="text-[10.5px] text-gray-500 leading-snug">{{ a.desc }}</span>
          </div>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />

    <div v-if="showNotes" class="hidden lg:block" />
    <div>
      <Connector label="6-stage pipeline" />
      <ArchBox icon="i-lucide-workflow" title="IngestBundleRunner" role="串行 per-connection" accent="violet">
        <div class="space-y-2">
          <div
            v-for="s in arch.stages"
            :key="s.name"
            class="rounded-xl border p-2.5"
            :class="ACCENTS[s.accent].surface"
          >
            <div class="flex items-center gap-1.5 mb-1">
              <div :class="[s.icon, ACCENTS[s.accent].text, 'text-sm']" />
              <span class="text-xs font-bold text-gray-700">{{ s.name }}</span>
              <span class="text-[10px] text-gray-400 ml-auto">{{ s.role }}</span>
            </div>
            <ul class="space-y-0.5">
              <li v-for="(b, i) in s.bullets" :key="i" class="text-[10.5px] text-gray-500 leading-snug flex gap-1.5">
                <span class="text-gray-300 flex-shrink-0">·</span>
                <span>{{ b }}</span>
              </li>
            </ul>
          </div>
        </div>
        <PeekPanel label="WU Agent 工具腰带" icon="i-lucide-wrench" :count="arch.workUnitTools.length" accent="violet" class="mt-2">
          <div class="space-y-1.5">
            <div v-for="t in arch.workUnitTools" :key="t.name" class="flex items-baseline gap-2">
              <code class="text-[10.5px] font-mono font-semibold text-violet-600 flex-shrink-0">{{ t.name }}</code>
              <span class="text-[10.5px] text-gray-500">{{ t.desc }}</span>
            </div>
          </div>
        </PeekPanel>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />

    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="indigo" :items="arch.insights.finalize" />
    </div>
    <div>
      <Connector label="commit artifacts" />
      <ArchBox icon="i-lucide-package-check" title="输出工件" accent="indigo">
        <div class="space-y-1.5 mb-2">
          <div v-for="a in arch.artifacts" :key="a.path" class="flex items-start gap-2">
            <div :class="[a.icon, 'text-indigo-500 text-xs mt-0.5 flex-shrink-0']" />
            <div>
              <code class="text-[10.5px] font-mono font-bold text-indigo-700">{{ a.path }}</code>
              <div class="text-[10.5px] text-gray-500">{{ a.desc }}</div>
            </div>
          </div>
        </div>
        <PeekPanel label="端口抽象" icon="i-lucide-plug" :count="arch.ports.length" accent="indigo">
          <div class="space-y-1.5">
            <div v-for="p in arch.ports" :key="p.name" class="flex items-baseline gap-2">
              <code class="text-[10.5px] font-mono font-semibold text-indigo-600 flex-shrink-0">{{ p.name }}</code>
              <span class="text-[10.5px] text-gray-500">{{ p.desc }}</span>
            </div>
          </div>
        </PeekPanel>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
