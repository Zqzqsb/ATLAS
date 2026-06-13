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
const arch = computed(() => getWrenModule(props.flow.id)?.skills ?? null)

const gridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 gap-x-6 gap-y-3 items-start lg:items-center" :class="gridCols">
    <!-- ════ Stage 1: BYO Agent ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="slate" :intro="arch.insights.input" />
    </div>
    <ArchBox icon="i-lucide-bot" :title="arch.input.label" accent="slate" muted>
      <div class="text-xs text-gray-500 leading-snug mb-2">{{ arch.input.note }}</div>
      <div class="flex flex-wrap gap-1.5">
        <span v-for="f in arch.input.frameworks" :key="f" class="px-2 py-0.5 rounded-full text-[11px] font-semibold bg-gray-100 text-gray-600 border border-gray-200">{{ f }}</span>
      </div>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- ════ Stage 2: Agent SDK (primitives as tools) ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="violet" :items="arch.insights.sdk" />
    </div>
    <div>
      <Connector label="import wren SDK" />
      <ArchBox icon="i-lucide-plug" :title="arch.sdk.title" role="function-calling" accent="violet">
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-violet-200 bg-violet-50/40 px-2.5 py-1.5 mb-2.5">
          <div class="i-lucide-info text-violet-400 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-violet-700 leading-snug">{{ arch.sdk.note }}</span>
        </div>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-1.5">
          <div v-for="t in arch.sdk.tools" :key="t.name" class="rounded-lg border border-violet-100 bg-white px-2 py-1.5">
            <code class="text-[11px] font-mono font-bold text-violet-700">{{ t.name }}</code>
            <div class="text-[10.5px] text-gray-500 leading-snug">{{ t.desc }}</div>
          </div>
        </div>
      </ArchBox>
    </div>
    <!-- right: `wren skills list` terminal -->
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-terminal text-slate-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">已安装 Skills</span>
          <span class="text-[11px] text-slate-400 ml-auto">可被 Agent 发现</span>
        </div>
        <pre class="text-[10.5px] leading-relaxed font-mono text-gray-600 bg-gray-900/95 text-gray-100 p-3 overflow-auto whitespace-pre">{{ arch.listCmd }}</pre>
      </div>
    </div>

    <!-- ════ Stage 3: Skills (markdown workflows) ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="slate" :items="arch.insights.skills" />
    </div>
    <div>
      <Connector label="按 skill 编排" />
      <ArchBox icon="i-lucide-scroll-text" title="Agent Skills" role="Markdown 工作流" accent="slate" :badge="`× ${arch.skills.length}`">
        <div class="space-y-2">
          <div v-for="s in arch.skills" :key="s.name" class="rounded-xl border p-2.5" :class="ACCENTS[s.accent].surface">
            <div class="flex items-center gap-1.5 mb-1">
              <div class="i-lucide-file-code text-sm flex-shrink-0" :class="ACCENTS[s.accent].text" />
              <code class="text-xs font-mono font-bold text-gray-800">{{ s.name }}</code>
              <span class="text-[10.5px] text-gray-400 ml-auto truncate">{{ s.when }}</span>
            </div>
            <PeekPanel :label="`工作流步骤`" icon="i-lucide-list-ordered" :count="s.steps.length" :accent="s.accent">
              <ol class="space-y-1 pl-0 list-none">
                <li v-for="(st, i) in s.steps" :key="i" class="flex items-start gap-2 text-[11px] text-gray-600 leading-relaxed">
                  <span class="w-3.5 h-3.5 rounded-full flex-center text-[8px] font-bold flex-shrink-0 mt-0.5" :class="[ACCENTS[s.accent].iconBg, ACCENTS[s.accent].text]">{{ i + 1 }}</span>
                  <span>{{ st }}</span>
                </li>
              </ol>
            </PeekPanel>
          </div>
        </div>
      </ArchBox>
    </div>
    <!-- right: sample skill markdown -->
    <div>
      <div class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100">
          <div class="i-lucide-file-text text-emerald-500 text-sm" />
          <span class="text-sm font-bold text-slate-800">Skill 示例</span>
          <span class="text-[11px] text-slate-400 ml-auto font-mono">{{ arch.sampleSkill.name }}</span>
        </div>
        <pre class="text-[10.5px] leading-relaxed font-mono text-gray-600 p-3 overflow-auto whitespace-pre-wrap">{{ arch.sampleSkill.md }}</pre>
      </div>
    </div>
  </div>
</template>
