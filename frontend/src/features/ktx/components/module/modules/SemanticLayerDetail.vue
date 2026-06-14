<script setup lang="ts">
import { computed } from 'vue'
import ArchBox from '../../../../arch/components/module/diagram/ArchBox.vue'
import Connector from '../../../../arch/components/module/diagram/Connector.vue'
import PeekPanel from '../../../../arch/components/module/diagram/PeekPanel.vue'
import InsightNotes from '../../../../arch/components/module/diagram/InsightNotes.vue'
import type { KtxFlowDef } from '../../../model/ktx'
import { getKtxModule } from '../../../model/ktx'

const props = defineProps<{ flow: KtxFlowDef; showNotes?: boolean }>()
const arch = computed(() => getKtxModule(props.flow.id)?.sl ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="amber" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-file-json" :title="arch.input.label" accent="slate" muted>
      <code class="text-[10.5px] font-mono text-gray-600">{{ arch.input.example }}</code>
      <div class="text-[11px] text-gray-500 mt-1">{{ arch.input.note }}</div>
    </ArchBox>
    <div class="hidden lg:block" />

    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="amber" :items="arch.insights.plan" />
    </div>
    <div>
      <Connector label="YAML sources" />
      <ArchBox icon="i-lucide-folder-open" :title="arch.loader.title" accent="emerald">
        <PeekPanel :label="arch.loader.title" icon="i-lucide-list" :count="arch.loader.items.length" accent="emerald">
          <div class="space-y-1.5">
            <div v-for="it in arch.loader.items" :key="it.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-emerald-600 flex-shrink-0">{{ it.name }}</code>
              <span class="text-[11px] text-gray-500">{{ it.desc }}</span>
            </div>
          </div>
        </PeekPanel>
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-blue-200 bg-blue-50/40 px-2.5 py-1.5 mt-2">
          <div class="i-lucide-git-merge text-blue-400 text-xs mt-0.5 flex-shrink-0" />
          <div class="text-[11px] text-blue-700 leading-snug">
            <strong>{{ arch.joinGraph.title }}</strong> — {{ arch.joinGraph.algo }}<br>
            {{ arch.joinGraph.note }}
          </div>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />

    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="rose" :items="arch.insights.fanout" />
    </div>
    <div>
      <Connector label="QueryPlanner.plan()" />
      <ArchBox icon="i-lucide-cpu" title="13-step Planner + Generator" role="python/ktx-sl" accent="amber" badge="× 8">
        <ol class="space-y-1 pl-0 list-none mb-2">
          <li v-for="(s, i) in arch.plannerSteps" :key="s.name" class="flex items-start gap-2 text-[11px] text-gray-600 leading-relaxed">
            <span class="w-3.5 h-3.5 rounded-full bg-amber-100 text-amber-700 flex-center text-[8px] font-bold flex-shrink-0 mt-0.5">{{ i + 1 }}</span>
            <span><code class="font-mono font-semibold text-amber-700">{{ s.name }}</code> · {{ s.desc }}</span>
          </li>
        </ol>
        <PeekPanel label="Generator 路径" icon="i-lucide-route" :count="arch.generator.paths.length" accent="amber" class="mb-2">
          <div class="space-y-1.5">
            <div v-for="p in arch.generator.paths" :key="p.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-amber-600 flex-shrink-0">{{ p.name }}</code>
              <span class="text-[11px] text-gray-500">{{ p.desc }}</span>
            </div>
            <div class="text-[10.5px] text-gray-500 border-t border-amber-100 pt-1.5">{{ arch.generator.transpile }}</div>
          </div>
        </PeekPanel>
        <PeekPanel label="validate 五项" icon="i-lucide-shield-check" :count="arch.validate.length" accent="rose">
          <div class="space-y-1.5">
            <div v-for="v in arch.validate" :key="v.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-rose-600 flex-shrink-0">{{ v.name }}</code>
              <span class="text-[11px] text-gray-500">{{ v.desc }}</span>
            </div>
          </div>
        </PeekPanel>
      </ArchBox>
    </div>
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-git-compare-arrows text-amber-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">声明式 → 方言 SQL</span>
        </div>
        <pre class="text-[10px] leading-relaxed font-mono text-gray-700 p-3 overflow-auto whitespace-pre">{{ arch.sqlBefore }}</pre>
        <div class="flex items-center justify-center py-0.5 border-t border-slate-100">
          <div class="i-lucide-chevron-down text-amber-300 text-sm" />
        </div>
        <pre class="text-[10px] leading-relaxed font-mono text-gray-500 bg-amber-50/40 p-3 overflow-auto whitespace-pre">{{ arch.sqlAfter }}</pre>
      </div>
    </div>
  </div>
</template>
