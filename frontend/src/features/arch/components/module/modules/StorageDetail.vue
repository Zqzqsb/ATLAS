<script setup lang="ts">
import { computed } from 'vue'
import { ACCENTS } from '../../../model/architecture'
import type { FlowDef } from '../../../model/flows'
import { getModule } from '../../../model/modules'
import ArchBox from '../diagram/ArchBox.vue'
import Connector from '../diagram/Connector.vue'
import PeekPanel from '../diagram/PeekPanel.vue'
import InsightNotes from '../diagram/InsightNotes.vue'

const props = defineProps<{ flow: FlowDef; showNotes?: boolean }>()
const arch = computed(() => getModule(props.flow.id)?.storage ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- ════ Stage 1: Datasource root ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="slate" :intro="arch.insights.root" />
    </div>
    <ArchBox icon="i-lucide-server" :title="arch.root.label" accent="slate" muted>
      <div class="flex items-center gap-2 mb-2">
        <code class="px-2 py-0.5 rounded-md bg-gray-900 text-gray-100 font-mono text-[11px]">{{ arch.root.table }}</code>
        <span class="text-[11px] text-gray-500 leading-snug">{{ arch.root.note }}</span>
      </div>
      <div class="flex items-start gap-2 rounded-lg border border-dashed border-slate-200 bg-slate-50/60 px-2.5 py-1.5">
        <div class="i-lucide-git-branch text-slate-400 text-xs mt-0.5 flex-shrink-0" />
        <span class="text-[11px] text-slate-600 leading-snug">{{ arch.root.cascade }}</span>
      </div>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- ════ Stage 2: Rich Context tables ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="indigo" :items="arch.insights.tables" />
    </div>
    <div>
      <Connector label="FK datasource_id" />
      <ArchBox icon="i-lucide-table-2" :title="arch.tables.title" :role="arch.tables.role" accent="indigo">
        <div class="space-y-1.5">
          <div v-for="t in arch.tables.items" :key="t.table" class="rounded-xl border p-2" :class="ACCENTS[t.accent].surface">
            <div class="flex items-center gap-2 mb-0.5">
              <code class="px-1.5 py-0.5 rounded bg-gray-900 text-gray-100 font-mono text-[11px] flex-shrink-0">{{ t.table }}</code>
              <span class="text-xs font-semibold text-gray-800">{{ t.label }}</span>
              <code v-if="t.flag" class="ml-auto px-1 rounded text-[9px] font-mono font-bold flex-shrink-0 bg-amber-50 text-amber-700 border border-amber-200">{{ t.flag }}</code>
            </div>
            <div class="text-[10.5px] text-gray-500 leading-snug font-mono">{{ t.cols }}</div>
          </div>
        </div>
      </ArchBox>
    </div>
    <!-- right: ER mini-map -->
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-network text-indigo-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">ER · 关系图</span>
          <span class="text-[11px] text-slate-400 ml-auto">ON DELETE CASCADE</span>
        </div>
        <div class="p-3">
          <div class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-slate-900 text-gray-100 font-mono text-[11px] font-bold mb-2">
            <div class="i-lucide-server text-xs" />{{ arch.er.root }}
          </div>
          <div class="pl-3 border-l-2 border-dashed border-indigo-200 space-y-1">
            <div v-for="e in arch.er.edges" :key="e.table" class="flex items-center gap-2">
              <div class="i-lucide-corner-down-right text-indigo-300 text-xs flex-shrink-0" />
              <code class="text-[11px] font-mono font-semibold text-indigo-700">{{ e.table }}</code>
              <span class="text-[10px] text-gray-400 ml-auto">{{ e.rel }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ════ Stage 3: Vector layer ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="indigo" :items="arch.insights.vector" />
    </div>
    <div>
      <Connector label="entity_type + entity_id" />
      <ArchBox icon="i-lucide-radar" :title="arch.vector.title" :role="arch.vector.role" accent="indigo">
        <div class="space-y-1 mb-2.5">
          <div v-for="s in arch.vector.spec" :key="s.name" class="flex items-baseline gap-2">
            <code class="text-[11px] font-mono font-semibold text-indigo-700 flex-shrink-0">{{ s.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug">{{ s.desc }}</span>
          </div>
        </div>
        <PeekPanel label="嵌入写入路径" icon="i-lucide-spline" :count="arch.embed.paths.length" accent="indigo">
          <div class="space-y-1.5">
            <div class="flex items-center gap-2 mb-1">
              <code class="text-[11px] font-mono font-bold text-indigo-700">{{ arch.embed.model }}</code>
              <span class="px-1.5 rounded text-[10px] font-bold bg-indigo-50 text-indigo-700 border border-indigo-200">{{ arch.embed.dim }}</span>
            </div>
            <div class="text-[11px] text-gray-500 leading-snug">{{ arch.embed.provider }}</div>
            <div v-for="(p, i) in arch.embed.paths" :key="i" class="flex items-start gap-1.5 text-[11px] text-gray-500 leading-snug">
              <div class="i-lucide-dot text-indigo-400 flex-shrink-0" />
              <span>{{ p }}</span>
            </div>
            <code class="block text-[10px] font-mono text-gray-400 mt-1">{{ arch.embed.upsert }}</code>
          </div>
        </PeekPanel>
      </ArchBox>
    </div>
    <!-- right: DDL + vector search SQL -->
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-file-code-2 text-indigo-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">向量召回 SQL</span>
          <span class="text-[11px] text-slate-400 ml-auto font-mono">VEC_DISTANCE_COSINE</span>
        </div>
        <pre class="text-[10px] leading-relaxed font-mono text-gray-700 p-3 overflow-auto whitespace-pre">{{ arch.vector.ddl }}</pre>
        <div class="flex items-center justify-center py-0.5 border-t border-slate-100">
          <div class="i-lucide-chevron-down text-indigo-300 text-sm" />
        </div>
        <pre class="text-[10px] leading-relaxed font-mono text-gray-500 bg-indigo-50/40 p-3 overflow-auto whitespace-pre">{{ arch.vector.search }}</pre>
        <div class="flex items-start gap-2 px-3 py-2 border-t border-slate-100 bg-amber-50/40">
          <div class="i-lucide-crosshair text-amber-500 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[10.5px] text-amber-700 leading-snug">{{ arch.vector.searchNote }}</span>
        </div>
      </div>
    </div>

    <!-- ════ Stage 4: Change log + flags ════ -->
    <div v-if="showNotes" class="hidden lg:block" />
    <div>
      <Connector label="审计 / 失效" />
      <ArchBox icon="i-lucide-scroll-text" :title="arch.changelog.table" role="自维护审计" accent="indigo">
        <div class="space-y-1.5 mb-2.5">
          <div v-for="c in arch.changelog.types" :key="c.name" class="flex items-baseline gap-2">
            <code class="text-[11px] font-mono font-semibold text-indigo-700 flex-shrink-0">{{ c.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug">{{ c.desc }}</span>
          </div>
        </div>
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-indigo-200 bg-indigo-50/40 px-2.5 py-1.5 mb-2.5">
          <div class="i-lucide-info text-indigo-400 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-indigo-700 leading-snug">{{ arch.changelog.note }}</span>
        </div>
        <PeekPanel label="三个失效标志位" icon="i-lucide-flag" :count="arch.flags.length" accent="amber">
          <div class="space-y-1.5">
            <div v-for="f in arch.flags" :key="f.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-amber-700 flex-shrink-0">{{ f.name }}</code>
              <span class="text-[11px] text-gray-500 leading-snug">{{ f.desc }}</span>
            </div>
          </div>
        </PeekPanel>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
