<script setup lang="ts">
import { computed } from 'vue'
import ArchBox from '../../../../arch/components/module/diagram/ArchBox.vue'
import Connector from '../../../../arch/components/module/diagram/Connector.vue'
import PeekPanel from '../../../../arch/components/module/diagram/PeekPanel.vue'
import InsightNotes from '../../../../arch/components/module/diagram/InsightNotes.vue'
import type { WrenFlowDef } from '../../../model/wren'
import { getWrenModule } from '../../../model/wren'

const props = defineProps<{ flow: WrenFlowDef; showNotes?: boolean }>()
const arch = computed(() => getWrenModule(props.flow.id)?.memory ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- ════ Stage 1: Input ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="slate" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-folder-git-2" :title="arch.input.label" accent="slate" muted>
      <div class="text-xs text-gray-500 leading-snug">{{ arch.input.note }}</div>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- ════ Stage 2: Index + Embed ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="blue" :items="arch.insights.index" />
    </div>
    <div>
      <Connector :label="arch.index.cmd" />
      <ArchBox icon="i-lucide-list-tree" title="Schema Indexer" role="工件 → 条目" accent="blue">
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-blue-200 bg-blue-50/40 px-2.5 py-1.5 mb-2.5">
          <div class="i-lucide-info text-blue-400 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-blue-700 leading-snug">{{ arch.index.note }}</span>
        </div>
        <PeekPanel label="索引条目类型" icon="i-lucide-boxes" :count="arch.index.items.length" accent="blue">
          <div class="space-y-1.5">
            <div v-for="it in arch.index.items" :key="it.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-blue-600 flex-shrink-0">{{ it.name }}</code>
              <span class="text-[11px] text-gray-500 leading-snug">{{ it.desc }}</span>
            </div>
          </div>
        </PeekPanel>
        <!-- embed config -->
        <div class="rounded-xl border border-blue-100 bg-white px-2.5 py-2 mt-2">
          <div class="flex items-center gap-1.5 mb-1">
            <div class="i-lucide-spline text-blue-500 text-sm" />
            <span class="text-[11px] font-bold text-gray-700">Embeddings</span>
            <code class="ml-auto px-1.5 rounded text-[10px] font-mono font-bold bg-blue-50 text-blue-700 border border-blue-200">{{ arch.embed.dim }}</code>
          </div>
          <code class="text-[10.5px] font-mono text-blue-700 break-all">{{ arch.embed.model }}</code>
          <div class="text-[10.5px] text-gray-500 leading-snug mt-0.5">{{ arch.embed.engine }} · {{ arch.embed.note }}</div>
        </div>
      </ArchBox>
    </div>
    <!-- right: LanceDB collections with live counts -->
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-database text-blue-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">LanceDB Collections</span>
          <span class="text-[11px] text-slate-400 ml-auto font-mono">~/.wren</span>
        </div>
        <div class="p-2.5 space-y-1.5">
          <div v-for="c in arch.collections" :key="c.table" class="rounded-xl border border-blue-100 bg-blue-50/40 px-2.5 py-2">
            <div class="flex items-center gap-1.5 mb-0.5">
              <div :class="[c.icon, 'text-blue-500 text-sm']" />
              <code class="text-[11px] font-mono font-bold text-blue-700">{{ c.table }}</code>
              <span class="ml-auto px-1.5 rounded-full text-[10px] font-bold bg-blue-100 text-blue-700">{{ c.count }}</span>
            </div>
            <div class="text-[10.5px] text-gray-500 leading-snug">{{ c.desc }}</div>
            <div class="text-[10px] text-blue-600/80 mt-0.5">用于 {{ c.use }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- ════ Stage 3: Retrieve ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="blue" :items="arch.insights.retrieve" />
    </div>
    <div>
      <Connector label="查询时召回" />
      <ArchBox icon="i-lucide-search" title="Retrieve" role="schema linking + few-shot" accent="blue">
        <div class="space-y-1.5 mb-2">
          <div class="rounded-lg bg-white border border-blue-100 px-2 py-1.5">
            <code class="text-[11px] font-mono font-bold text-blue-700">{{ arch.retrieve.fetch.split(' — ')[0] }}</code>
            <div class="text-[11px] text-gray-500 leading-snug">{{ arch.retrieve.fetch.split(' — ')[1] }}</div>
          </div>
          <div class="rounded-lg bg-white border border-blue-100 px-2 py-1.5">
            <code class="text-[11px] font-mono font-bold text-blue-700">{{ arch.retrieve.recall.split(' — ')[0] }}</code>
            <div class="text-[11px] text-gray-500 leading-snug">{{ arch.retrieve.recall.split(' — ')[1] }}</div>
          </div>
        </div>
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-blue-200 bg-blue-50/40 px-2.5 py-1.5">
          <div class="i-lucide-layers text-blue-400 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-blue-700 leading-snug">{{ arch.retrieve.strategy }}</span>
        </div>
      </ArchBox>
    </div>
    <!-- right: query_history sample pairs (mocked) -->
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-history text-blue-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">query_history 样例</span>
          <span class="text-[11px] text-slate-400 ml-auto">few-shot 召回源</span>
        </div>
        <div class="p-2.5 space-y-2">
          <div v-for="s in arch.samples" :key="s.nl" class="rounded-xl border border-slate-100 bg-slate-50/50 overflow-hidden">
            <div class="flex items-center gap-1.5 px-2.5 py-1.5 border-b border-slate-100">
              <div class="i-lucide-message-square text-blue-400 text-xs flex-shrink-0" />
              <span class="text-[11px] font-semibold text-gray-700">{{ s.nl }}</span>
              <code
                class="ml-auto px-1 rounded text-[9px] font-mono font-bold flex-shrink-0"
                :class="s.tag === 'confirmed' ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-amber-50 text-amber-700 border border-amber-200'"
              >{{ s.tag }}</code>
            </div>
            <pre class="text-[10px] leading-relaxed font-mono text-gray-500 px-2.5 py-1.5 overflow-auto whitespace-pre">{{ s.sql }}</pre>
          </div>
        </div>
      </div>
    </div>

    <!-- ════ Stage 4: Seed / Store lifecycle (closed loop) ════ -->
    <div v-if="showNotes" class="hidden lg:block" />
    <div>
      <Connector label="冷启动 ↔ 沉淀" />
      <ArchBox icon="i-lucide-recycle" title="Seed / Store 生命周期" role="使用即沉淀" accent="emerald">
        <div class="space-y-1.5">
          <div class="rounded-xl border border-emerald-100 bg-emerald-50/40 px-2.5 py-1.5">
            <code class="text-[11px] font-mono font-bold text-emerald-700">{{ arch.seed.cmd }}</code>
            <div class="text-[11px] text-gray-500 leading-snug">{{ arch.seed.note }}</div>
          </div>
          <div class="rounded-xl border border-emerald-100 bg-emerald-50/40 px-2.5 py-1.5">
            <code class="text-[11px] font-mono font-bold text-emerald-700">{{ arch.store.cmd }}</code>
            <div class="text-[11px] text-gray-500 leading-snug">{{ arch.store.note }}</div>
          </div>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
