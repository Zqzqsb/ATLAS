<script setup lang="ts">
import { computed } from 'vue'
import ArchBox from '../../../../arch/components/module/diagram/ArchBox.vue'
import Connector from '../../../../arch/components/module/diagram/Connector.vue'
import PeekPanel from '../../../../arch/components/module/diagram/PeekPanel.vue'
import InsightNotes from '../../../../arch/components/module/diagram/InsightNotes.vue'
import type { WrenFlowDef } from '../../../model/wren'
import { getWrenModule } from '../../../model/wren'

const props = defineProps<{ flow: WrenFlowDef; showNotes?: boolean }>()
const arch = computed(() => getWrenModule(props.flow.id)?.execution ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- ════ Stage 1: Executable SQL ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="slate" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-file-code-2" :title="arch.input.label" accent="slate" muted>
      <div class="text-xs text-gray-500 leading-snug">{{ arch.input.note }}</div>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- ════ Stage 2: Connect (profiles + connectors) ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="indigo" :items="arch.insights.connect" />
    </div>
    <div>
      <Connector label="按 profile 连接" />
      <ArchBox icon="i-lucide-plug" title="Connectors" role="原生 · 20+ 源" accent="indigo">
        <!-- grouped data sources -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-1.5 mb-2.5">
          <div v-for="c in arch.connectors" :key="c.group" class="rounded-xl border border-indigo-100 bg-indigo-50/40 px-2.5 py-2">
            <div class="flex items-center gap-1.5 mb-0.5">
              <div :class="[c.icon, 'text-indigo-500 text-xs flex-shrink-0']" />
              <span class="text-[11px] font-bold text-gray-700">{{ c.group }}</span>
            </div>
            <div class="text-[10.5px] text-gray-500 leading-snug">{{ c.items }}</div>
          </div>
        </div>
        <!-- credential separation -->
        <PeekPanel :label="arch.profiles.title" icon="i-lucide-key-round" :count="arch.profiles.fields.length" accent="indigo">
          <div class="space-y-1.5">
            <p class="text-[11px] text-gray-500 leading-snug mb-1">{{ arch.profiles.note }}</p>
            <div v-for="f in arch.profiles.fields" :key="f.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-indigo-600 flex-shrink-0">{{ f.name }}</code>
              <span class="text-[11px] text-gray-500 leading-snug">{{ f.desc }}</span>
            </div>
          </div>
        </PeekPanel>
      </ArchBox>
    </div>
    <!-- right: profiles.yml example -->
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-key-round text-indigo-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">凭据与项目分离</span>
          <span class="text-[11px] text-slate-400 ml-auto font-mono">~/.wren</span>
        </div>
        <pre class="text-[10.5px] leading-relaxed font-mono text-gray-600 p-3 overflow-auto whitespace-pre">{{ arch.profiles.example }}</pre>
      </div>
    </div>

    <!-- ════ Stage 3: Dry-run → Execute → Result ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="indigo" :items="arch.insights.execute" />
    </div>
    <div>
      <Connector :label="arch.dryrun.cmd" />
      <ArchBox icon="i-lucide-flask-conical" title="Dry-run → Execute" role="先校验后执行" accent="indigo">
        <ul class="space-y-1 mb-2.5">
          <li v-for="(p, i) in arch.dryrun.points" :key="i" class="flex items-start gap-2 text-[11px] text-gray-600 leading-relaxed">
            <div class="i-lucide-check mt-0.5 flex-shrink-0 text-indigo-500" />
            <span>{{ p }}</span>
          </li>
        </ul>
        <div class="rounded-xl border border-indigo-100 bg-white px-2.5 py-2">
          <div class="flex items-center gap-1.5 mb-1">
            <div class="i-lucide-table text-indigo-500 text-sm" />
            <span class="text-[11px] font-bold text-gray-700">结果格式</span>
          </div>
          <div class="space-y-1">
            <div v-for="f in arch.result.formats" :key="f.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-indigo-600 flex-shrink-0 w-24">{{ f.name }}</code>
              <span class="text-[11px] text-gray-500 leading-snug">{{ f.desc }}</span>
            </div>
          </div>
        </div>
      </ArchBox>
    </div>
    <!-- right: result table preview (mocked) -->
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-table text-indigo-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">PyArrow 结果</span>
          <span class="text-[11px] text-slate-400 ml-auto font-mono">limit 100</span>
        </div>
        <table class="w-full text-[11px]">
          <thead>
            <tr class="bg-indigo-50/50 text-indigo-700">
              <th v-for="col in arch.result.preview.cols" :key="col" class="text-left font-mono font-bold px-3 py-1.5">{{ col }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, ri) in arch.result.preview.rows" :key="ri" class="border-t border-slate-100">
              <td v-for="(cell, ci) in row" :key="ci" class="px-3 py-1.5 font-mono text-gray-600">{{ cell }}</td>
            </tr>
          </tbody>
        </table>
        <div class="text-[10px] text-gray-400 px-3 py-1.5 border-t border-slate-100">… 4 / 100 rows</div>
      </div>
    </div>
  </div>
</template>
