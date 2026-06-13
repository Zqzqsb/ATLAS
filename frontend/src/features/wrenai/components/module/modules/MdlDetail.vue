<script setup lang="ts">
import { computed } from 'vue'
import { ACCENTS } from '../../../../arch/model/architecture'
import ArchBox from '../../../../arch/components/module/diagram/ArchBox.vue'
import Connector from '../../../../arch/components/module/diagram/Connector.vue'
import PeekPanel from '../../../../arch/components/module/diagram/PeekPanel.vue'
import InsightNotes from '../../../../arch/components/module/diagram/InsightNotes.vue'
import type { WrenFlowDef } from '../../../model/wren'
import { getWrenModule } from '../../../model/wren'

const props = defineProps<{ flow: WrenFlowDef; showNotes?: boolean }>()
const arch = computed(() => getWrenModule(props.flow.id)?.mdl ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- ════ Stage 1: Sourcing (where MDL comes from) ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="emerald" :intro="arch.insights.sourcing" />
    </div>
    <ArchBox icon="i-lucide-git-fork" :title="arch.sourcing.title" accent="emerald" badge="× 6">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-1.5">
        <div
          v-for="p in arch.sourcing.paths"
          :key="p.name"
          class="rounded-xl border p-2 flex flex-col gap-1"
          :class="ACCENTS[p.accent].surface"
        >
          <div class="flex items-center gap-1.5">
            <div :class="[p.icon, ACCENTS[p.accent].text, 'text-sm flex-shrink-0']" />
            <span class="text-xs font-bold text-gray-800 leading-tight">{{ p.name }}</span>
            <code v-if="p.badge" class="ml-auto px-1 rounded text-[9px] font-mono font-bold flex-shrink-0" :class="ACCENTS[p.accent].chip">{{ p.badge }}</code>
          </div>
          <span class="text-[11px] text-gray-500 leading-snug">{{ p.desc }}</span>
        </div>
      </div>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- ════ Stage 2: Modeling (MDL Manifest) ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="emerald" :items="arch.insights.model" />
    </div>
    <div>
      <Connector label="统一为项目 YAML" />
      <ArchBox icon="i-lucide-box" title="MDL Manifest" role="语义契约" accent="emerald" :badge="arch.project.name">
        <!-- demo project identity (mocked from jaffle_shop) -->
        <div class="flex items-center gap-2 rounded-lg border border-emerald-200 bg-emerald-50/50 px-2.5 py-1.5 mb-2.5">
          <div class="i-lucide-folder-git-2 text-emerald-500 text-sm flex-shrink-0" />
          <div class="min-w-0">
            <div class="text-[11px] font-bold text-emerald-700 font-mono">{{ arch.project.stats }}</div>
            <div class="text-[10.5px] text-gray-500 leading-snug truncate">{{ arch.project.desc }}</div>
          </div>
        </div>
        <!-- modeled entities -->
        <div class="space-y-2 mb-2.5">
          <div v-for="e in arch.entities" :key="e.title" class="rounded-xl border p-2.5" :class="ACCENTS[e.accent].surface">
            <div class="flex items-center gap-1.5 mb-1">
              <div :class="[e.icon, ACCENTS[e.accent].text, 'text-sm']" />
              <span class="text-xs font-bold text-gray-700">{{ e.title }}</span>
              <span class="text-[11px] text-gray-400 ml-1 truncate">{{ e.desc }}</span>
            </div>
            <PeekPanel :label="`${e.title} 字段`" icon="i-lucide-list" :count="e.items.length" :accent="e.accent">
              <div class="space-y-1.5">
                <div v-for="it in e.items" :key="it.name" class="flex items-baseline gap-2">
                  <code class="text-[11px] font-mono font-semibold flex-shrink-0" :class="ACCENTS[e.accent].text">{{ it.name }}</code>
                  <span class="text-[11px] text-gray-500 leading-snug">{{ it.desc }}</span>
                </div>
              </div>
            </PeekPanel>
          </div>
        </div>

        <!-- governance baked into model -->
        <div class="rounded-xl border border-rose-200 bg-rose-50/40 p-2.5">
          <div class="flex items-center gap-1.5 mb-1.5">
            <div class="i-lucide-shield text-rose-500 text-sm" />
            <span class="text-xs font-bold text-gray-700">{{ arch.govern.title }}</span>
          </div>
          <div class="space-y-1.5">
            <div v-for="g in arch.govern.items" :key="g.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-rose-600 flex-shrink-0">{{ g.name }}</code>
              <span class="text-[11px] text-gray-500 leading-snug">{{ g.desc }}</span>
            </div>
          </div>
        </div>
      </ArchBox>
    </div>
    <!-- right: YAML → JSON illustration + 5-layer context -->
    <div class="space-y-3">
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-file-code-2 text-emerald-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">模型定义</span>
          <span class="text-[11px] text-slate-400 ml-auto font-mono">customers (YAML)</span>
        </div>
        <pre class="text-[10.5px] leading-relaxed font-mono text-gray-700 p-3 overflow-auto whitespace-pre">{{ arch.yamlExample }}</pre>
        <div class="flex items-center justify-center py-0.5 border-t border-slate-100">
          <div class="i-lucide-chevron-down text-gray-300 text-sm" />
        </div>
        <pre class="text-[10.5px] leading-relaxed font-mono text-gray-500 bg-gray-50/60 p-3 overflow-auto whitespace-pre">{{ arch.jsonExample }}</pre>
      </div>
      <!-- 5-layer context model -->
      <div class="rounded-2xl border border-slate-200 bg-white px-3.5 py-3">
        <div class="flex items-center gap-1.5 mb-2">
          <div class="i-lucide-layers text-violet-500 text-sm" />
          <span class="text-xs font-bold text-slate-700">五层上下文模型</span>
          <span class="text-[10px] text-slate-400 ml-auto">MDL 承载前 3 层</span>
        </div>
        <div class="space-y-1">
          <div
            v-for="(c, i) in arch.contextLayers"
            :key="c.name"
            class="flex items-baseline gap-2 px-2 py-1 rounded-md"
            :class="i < 3 ? 'bg-emerald-50/60' : 'bg-gray-50'"
          >
            <code class="text-[11px] font-mono font-semibold flex-shrink-0 w-20" :class="i < 3 ? 'text-emerald-700' : 'text-gray-400'">{{ c.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug">{{ c.desc }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- ════ Stage 3: Compile ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="indigo" :items="arch.insights.compile" />
    </div>
    <div>
      <Connector :label="arch.compile.cmd" />
      <ArchBox icon="i-lucide-package-check" title="Compile → Manifest" accent="indigo">
        <div class="flex items-center gap-2 mb-2">
          <code class="px-2 py-0.5 rounded-md bg-gray-900 text-gray-100 font-mono text-[11px]">{{ arch.compile.cmd }}</code>
          <div class="i-lucide-arrow-right text-gray-300 text-xs" />
          <code class="px-2 py-0.5 rounded-md bg-indigo-100 text-indigo-700 font-mono text-[11px]">{{ arch.compile.out }}</code>
        </div>
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-indigo-200 bg-indigo-50/40 px-2.5 py-1.5">
          <div class="i-lucide-info text-indigo-400 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-indigo-700 leading-snug">{{ arch.compile.note }}</span>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
