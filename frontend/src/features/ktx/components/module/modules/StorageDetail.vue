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
const arch = computed(() => getKtxModule(props.flow.id)?.storage ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="indigo" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-folder-git-2" :title="arch.input.label" accent="slate" muted>
      <div class="text-[11px] text-gray-500 leading-snug">{{ arch.input.note }}</div>
    </ArchBox>
    <div class="hidden lg:block" />

    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="emerald" :items="arch.insights.git" />
    </div>
    <div>
      <Connector label="ktx setup" />
      <ArchBox icon="i-lucide-folder-tree" title="Project Layout" accent="indigo" badge="× 6">
        <div class="space-y-1.5">
          <div
            v-for="p in arch.paths"
            :key="p.path"
            class="rounded-xl border p-2.5"
            :class="ACCENTS[p.accent].surface"
          >
            <div class="flex items-center gap-1.5 mb-0.5">
              <div :class="[p.icon, ACCENTS[p.accent].text, 'text-sm']" />
              <code class="text-[11px] font-mono font-bold text-gray-800">{{ p.path }}</code>
              <span
                class="ml-auto text-[10px] font-bold px-1.5 rounded"
                :class="p.commit === '✓' ? 'bg-emerald-100 text-emerald-700' : 'bg-gray-100 text-gray-500'"
              >git {{ p.commit }}</span>
            </div>
            <div class="text-[10px] font-semibold text-gray-500">{{ p.role }}</div>
            <div class="text-[10.5px] text-gray-500">{{ p.desc }}</div>
          </div>
        </div>
      </ArchBox>
    </div>
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-settings text-slate-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">ktx.yaml</span>
          <span class="text-[11px] text-emerald-600 ml-auto font-semibold">commit ✓</span>
        </div>
        <pre class="text-[10px] leading-relaxed font-mono text-gray-600 p-3 overflow-auto whitespace-pre">{{ arch.ktxYaml }}</pre>
      </div>
    </div>

    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="indigo" :items="arch.insights.state" />
    </div>
    <div>
      <Connector label=".ktx/ local state" />
      <ArchBox icon="i-lucide-database" title=".ktx/db.sqlite" accent="indigo">
        <PeekPanel label="索引表" icon="i-lucide-table-2" :count="arch.sqliteTables.length" accent="indigo" class="mb-2">
          <div class="space-y-1.5">
            <div v-for="t in arch.sqliteTables" :key="t.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-indigo-600 flex-shrink-0">{{ t.name }}</code>
              <span class="text-[11px] text-gray-500">{{ t.desc }}</span>
            </div>
          </div>
        </PeekPanel>
        <div class="rounded-xl border border-slate-200 bg-slate-50/50 px-2.5 py-2">
          <div class="text-xs font-bold text-gray-700 mb-1">项目解析顺序</div>
          <div v-for="r in arch.resolution" :key="r.name" class="flex items-baseline gap-2">
            <code class="text-[10.5px] font-mono font-semibold text-gray-600 flex-shrink-0">{{ r.name }}</code>
            <span class="text-[10.5px] text-gray-500">{{ r.desc }}</span>
          </div>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
