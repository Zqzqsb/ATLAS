<script setup lang="ts">
import { computed } from 'vue'
import { ACCENTS } from '../../../model/architecture'
import type { FlowDef } from '../../../model/flows'
import { getModule } from '../../../model/modules'
import ArchBox from '../diagram/ArchBox.vue'
import Connector from '../diagram/Connector.vue'
import PeekPanel from '../diagram/PeekPanel.vue'
import HealLoopDemo from '../diagram/HealLoopDemo.vue'
import InsightNotes from '../diagram/InsightNotes.vue'

const props = defineProps<{ flow: FlowDef; showNotes?: boolean }>()
const arch = computed(() => getModule(props.flow.id)?.maintain ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- ════ Stage 1: Trigger / Signal ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="slate" :intro="arch.insights.trigger" />
    </div>
    <ArchBox icon="i-lucide-git-pull-request-arrow" :title="arch.trigger.label" accent="slate" muted>
      <div class="flex items-center gap-2 text-xs text-gray-500 mb-2">
        <code class="px-1.5 py-0.5 rounded bg-gray-100 text-gray-600 font-mono text-[11px]">{{ arch.trigger.note }}</code>
      </div>
      <PeekPanel label="识别的变更类型" icon="i-lucide-list-tree" :count="arch.trigger.changeTypes.length" accent="slate">
        <div class="space-y-1.5">
          <div v-for="c in arch.trigger.changeTypes" :key="c.name" class="flex items-baseline gap-2">
            <code class="text-[11px] font-mono font-semibold text-slate-700 flex-shrink-0">{{ c.name }}</code>
            <span class="text-[11px] text-gray-500 leading-snug">{{ c.desc }}</span>
          </div>
        </div>
      </PeekPanel>
      <div class="flex items-start gap-2 rounded-lg border border-dashed border-slate-200 bg-slate-50/60 px-2.5 py-1.5 mt-2">
        <div class="i-lucide-info text-slate-400 text-xs mt-0.5 flex-shrink-0" />
        <span class="text-[11px] text-slate-600 leading-snug">{{ arch.trigger.detect }}</span>
      </div>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- ════ Stage 2: Coordinator + Executor (merged) ════ -->
    <div v-if="showNotes" class="hidden lg:block space-y-2">
      <InsightNotes accent="violet" :items="arch.insights.coordinator" />
      <InsightNotes accent="emerald" :items="arch.insights.executor" />
    </div>
    <div>
      <Connector />
      <!-- Coordinator -->
      <ArchBox icon="i-lucide-split" :title="arch.coordinator.title" :role="arch.coordinator.role" accent="violet">
        <div class="text-[10px] text-gray-400 font-mono mb-1.5">{{ arch.coordinator.engine }}</div>
        <ul class="space-y-1 mb-2">
          <li v-for="(p, i) in arch.coordinator.points" :key="i" class="flex items-start gap-2 text-xs text-gray-600 leading-relaxed">
            <div class="i-lucide-check mt-0.5 flex-shrink-0 text-violet-500" />
            <span>{{ p }}</span>
          </li>
        </ul>
        <PeekPanel label="Coordinator Tools" icon="i-lucide-wrench" :count="arch.coordinator.tools.length" accent="violet">
          <div class="space-y-1.5">
            <div v-for="t in arch.coordinator.tools" :key="t.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-violet-700 flex-shrink-0">{{ t.name }}</code>
              <span class="text-[11px] text-gray-500 leading-snug">{{ t.desc }}</span>
            </div>
          </div>
        </PeekPanel>
        <div class="rounded-xl border border-violet-200 bg-violet-50/40 px-2.5 py-2 mt-2.5">
          <div class="flex items-center gap-1.5 mb-1">
            <div class="i-lucide-arrow-down-to-line text-violet-600 text-sm" />
            <span class="text-xs font-bold text-gray-700">OUTPUT · {{ arch.coordinator.output.label }}</span>
          </div>
          <div class="flex flex-wrap gap-1">
            <span
              v-for="p in arch.coordinator.output.parts"
              :key="p"
              class="px-1.5 py-0.5 rounded-md bg-white border border-violet-200 text-[11px] text-violet-700"
            >{{ p }}</span>
          </div>
        </div>
      </ArchBox>

      <Connector :label="arch.executor.dispatch" />

      <!-- Executor -->
      <ArchBox icon="i-lucide-wand-2" :title="arch.executor.title" :role="arch.executor.role" accent="emerald">
        <div class="text-[10px] text-gray-400 font-mono mb-1.5">{{ arch.executor.engine }}</div>
        <ol class="space-y-1 mb-2 pl-0 list-none">
          <li v-for="(s, i) in arch.executor.steps" :key="i" class="flex items-start gap-2 text-xs text-gray-600 leading-relaxed">
            <span class="w-3.5 h-3.5 rounded-full bg-emerald-100 text-emerald-700 flex-center text-[8px] font-bold flex-shrink-0 mt-0.5">{{ i + 1 }}</span>
            <span>{{ s }}</span>
          </li>
        </ol>
        <PeekPanel label="Executor Tools" icon="i-lucide-wrench" :count="arch.executor.tools.length" accent="emerald">
          <div class="space-y-1.5">
            <div v-for="t in arch.executor.tools" :key="t.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-emerald-700 flex-shrink-0">{{ t.name }}</code>
              <span class="text-[11px] text-gray-500 leading-snug">{{ t.desc }}</span>
            </div>
          </div>
        </PeekPanel>
        <div class="flex items-center justify-center gap-2 mt-2.5 mb-2">
          <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-emerald-50 border border-emerald-200 text-[11px] font-semibold text-emerald-700">
            <div class="i-lucide-gauge text-xs" />{{ arch.executor.budget }}
          </div>
        </div>
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-emerald-200 bg-emerald-50/40 px-2.5 py-1.5">
          <div class="i-lucide-skip-forward text-emerald-500 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-emerald-700 leading-snug">{{ arch.executor.note }}</span>
        </div>
      </ArchBox>
    </div>
    <!-- right: heal-loop demo (aligned to Coordinator↔Executor) -->
    <div>
      <HealLoopDemo />
    </div>

    <!-- ════ Stage 3: Invalidation & Re-embed ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="indigo" :items="arch.insights.storage" />
    </div>
    <div>
      <Connector label="set_rich_context / delete" />
      <ArchBox icon="i-lucide-database" :title="arch.storage.title" accent="indigo">
        <div class="space-y-1.5 mb-2.5">
          <div v-for="item in arch.storage.items" :key="item.table" class="flex items-center gap-2.5">
            <code class="px-2 py-0.5 rounded-md bg-gray-900 text-gray-100 font-mono text-[11px] flex-shrink-0">{{ item.table }}</code>
            <div class="flex-1 min-w-0">
              <div class="text-xs font-semibold text-gray-800">{{ item.label }}</div>
              <div class="text-[11px] text-gray-400 truncate">{{ item.note }}</div>
            </div>
            <code class="hidden md:block text-[10px] font-mono text-gray-400 flex-shrink-0">{{ item.spec }}</code>
          </div>
        </div>
        <PeekPanel label="三个失效标志位" icon="i-lucide-flag" :count="arch.storage.flags.length" accent="indigo">
          <div class="space-y-1.5">
            <div v-for="f in arch.storage.flags" :key="f.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-indigo-700 flex-shrink-0">{{ f.name }}</code>
              <span class="text-[11px] text-gray-500 leading-snug">{{ f.desc }}</span>
            </div>
          </div>
        </PeekPanel>
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-indigo-200 bg-indigo-50/40 px-2.5 py-1.5 mt-2">
          <div class="i-lucide-radar text-indigo-400 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-indigo-700 leading-snug">{{ arch.storage.embed }}</span>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
