<script setup lang="ts">
import { computed } from 'vue'
import ArchBox from '../../../../arch/components/module/diagram/ArchBox.vue'
import Connector from '../../../../arch/components/module/diagram/Connector.vue'
import PeekPanel from '../../../../arch/components/module/diagram/PeekPanel.vue'
import InsightNotes from '../../../../arch/components/module/diagram/InsightNotes.vue'
import type { KtxFlowDef } from '../../../model/ktx'
import { getKtxModule } from '../../../model/ktx'

const props = defineProps<{ flow: KtxFlowDef; showNotes?: boolean }>()
const arch = computed(() => getKtxModule(props.flow.id)?.memory ?? null)

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
    <ArchBox icon="i-lucide-message-square" :title="arch.input.label" accent="slate" muted>
      <div class="flex flex-wrap gap-1 mb-1.5">
        <span
          v-for="s in arch.input.sources"
          :key="s"
          class="px-2 py-0.5 rounded-full text-[10px] font-semibold bg-slate-100 text-slate-600 border border-slate-200"
        >{{ s }}</span>
      </div>
      <div class="text-[11px] text-gray-500">{{ arch.input.note }}</div>
    </ArchBox>
    <div class="hidden lg:block" />

    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="violet" :items="arch.insights.isolation" />
    </div>
    <div>
      <Connector label="detectCaptureSignals" />
      <ArchBox icon="i-lucide-radar" title="Capture Signals" role="预过滤" accent="amber">
        <div class="space-y-1.5">
          <div v-for="s in arch.signals" :key="s.name" class="flex items-baseline gap-2">
            <code class="text-[11px] font-mono font-semibold text-amber-700 flex-shrink-0">{{ s.name }}</code>
            <span class="text-[11px] text-gray-500">{{ s.desc }}</span>
          </div>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />

    <div v-if="showNotes" class="hidden lg:block" />
    <div>
      <Connector label="worktree 隔离" />
      <ArchBox icon="i-lucide-brain-circuit" title="MemoryAgentService.ingest()" role="异步后台" accent="violet">
        <div class="space-y-1.5 mb-2">
          <div v-for="(w, i) in arch.worktreeFlow" :key="w.name" class="flex items-start gap-2">
            <span class="w-4 h-4 rounded-full bg-violet-100 text-violet-700 flex-center text-[9px] font-bold flex-shrink-0 mt-0.5">{{ i + 1 }}</span>
            <div>
              <code class="text-[11px] font-mono font-bold text-violet-700">{{ w.name }}</code>
              <div class="text-[10.5px] text-gray-500">{{ w.desc }}</div>
            </div>
          </div>
        </div>
        <PeekPanel label="Agent 工具腰带" icon="i-lucide-wrench" :count="arch.toolBelt.length" accent="violet" class="mb-2">
          <div class="space-y-1.5">
            <div v-for="t in arch.toolBelt" :key="t.name" class="flex items-baseline gap-2">
              <code class="text-[10.5px] font-mono font-semibold text-violet-600 flex-shrink-0">{{ t.name }}</code>
              <span class="text-[10.5px] text-gray-500">{{ t.desc }}</span>
            </div>
          </div>
        </PeekPanel>
        <div class="rounded-xl border border-rose-200 bg-rose-50/40 px-2.5 py-2">
          <div class="text-xs font-bold text-gray-700 mb-1">{{ arch.validation.title }}</div>
          <div v-for="v in arch.validation.items" :key="v.name" class="flex items-baseline gap-2 mb-0.5">
            <code class="text-[10.5px] font-mono font-semibold text-rose-600 flex-shrink-0">{{ v.name }}</code>
            <span class="text-[10.5px] text-gray-500">{{ v.desc }}</span>
          </div>
          <div class="text-[10.5px] text-rose-700 border-t border-rose-100 pt-1 mt-1">{{ arch.validation.note }}</div>
        </div>
      </ArchBox>
    </div>
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-clipboard-list text-violet-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">MemoryRunRecord</span>
        </div>
        <div class="p-2.5 space-y-1">
          <div v-for="r in arch.resultRecord" :key="r.field" class="flex items-baseline gap-2 px-2 py-1 rounded-md bg-violet-50/40">
            <code class="text-[10.5px] font-mono font-bold text-violet-700 flex-shrink-0 w-28">{{ r.field }}</code>
            <span class="text-[10.5px] text-gray-500">{{ r.desc }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="emerald" :items="arch.insights.learn" />
    </div>
    <div>
      <Connector label="squash → main" />
      <ArchBox icon="i-lucide-recycle" title="自改进闭环" accent="emerald">
        <div class="text-[11px] text-gray-600 leading-relaxed">
          对话尾学到的事实 → wiki Markdown / SL YAML → git commit → 下次 search/ingest 召回 → Agent 更准确
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />
  </div>
</template>
