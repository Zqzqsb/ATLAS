<script setup lang="ts">
import { computed } from 'vue'
import { ACCENTS } from '../../../model/architecture'
import type { FlowDef } from '../../../model/flows'
import { getModule } from '../../../model/modules'
import ArchBox from '../diagram/ArchBox.vue'
import Connector from '../diagram/Connector.vue'
import PeekPanel from '../diagram/PeekPanel.vue'
import LinkingDemo from '../diagram/LinkingDemo.vue'
import InsightNotes from '../diagram/InsightNotes.vue'

const props = defineProps<{ flow: FlowDef; showNotes?: boolean }>()
const arch = computed(() => getModule(props.flow.id)?.inference ?? null)

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
    <ArchBox icon="i-lucide-message-square" :title="arch.input.label" accent="slate" muted>
      <div class="flex items-center gap-2 text-xs text-gray-500">
        <code class="px-1.5 py-0.5 rounded bg-gray-100 text-gray-600 font-mono text-[11px]">{{ arch.input.example }}</code>
        <span>{{ arch.input.note }}</span>
      </div>
    </ArchBox>
    <div class="hidden lg:block" />

    <!-- ════ Stage 2: Adaptive Grounding ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="violet" :items="arch.insights.grounding" />
    </div>
    <div>
      <Connector />
      <ArchBox icon="i-lucide-search" :title="arch.grounding.title" :role="arch.grounding.role" accent="violet">
        <!-- Strategy dispatcher -->
        <ul class="space-y-1 mb-2">
          <li v-for="(p, i) in arch.grounding.dispatcher.points" :key="i" class="flex items-start gap-2 text-xs text-gray-600 leading-relaxed">
            <div class="i-lucide-check mt-0.5 flex-shrink-0 text-violet-500" />
            <span>{{ p }}</span>
          </li>
        </ul>
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-violet-200 bg-violet-50/40 px-2.5 py-1.5 mb-2.5">
          <div class="i-lucide-info text-violet-400 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-violet-700 leading-snug">{{ arch.grounding.dispatcher.note }}</span>
        </div>

        <!-- Retriever + Agent (two collaborators) -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-2.5 mb-2.5">
          <!-- CoarseRetriever -->
          <div class="rounded-xl border p-2.5" :class="ACCENTS.blue.surface">
            <div class="flex items-center gap-1.5 mb-1">
              <div class="i-lucide-radar text-blue-600 text-sm" />
              <span class="text-xs font-bold text-gray-700">{{ arch.grounding.retriever.title }}</span>
            </div>
            <p class="text-[11px] text-gray-500 leading-snug mb-2">{{ arch.grounding.retriever.desc }}</p>
            <PeekPanel label="4 路并发召回信号" icon="i-lucide-git-fork" :count="arch.grounding.retriever.signals.length" accent="blue">
              <div class="space-y-1.5">
                <div v-for="s in arch.grounding.retriever.signals" :key="s.name" class="flex items-baseline gap-2">
                  <code class="text-[11px] font-mono font-semibold text-blue-700 flex-shrink-0">{{ s.name }}</code>
                  <span class="text-[11px] text-gray-500 leading-snug">{{ s.desc }}</span>
                </div>
              </div>
            </PeekPanel>
          </div>

          <!-- LinkingAgent -->
          <div class="rounded-xl border p-2.5" :class="ACCENTS.violet.surface">
            <div class="flex items-center gap-1.5 mb-1">
              <div class="i-lucide-brain text-violet-600 text-sm" />
              <span class="text-xs font-bold text-gray-700">{{ arch.grounding.agent.title }}</span>
              <span class="text-[10px] text-gray-400 font-mono ml-auto">{{ arch.grounding.agent.engine }}</span>
            </div>
            <PeekPanel label="Linking 三种模式" icon="i-lucide-toggle-right" :count="arch.grounding.agent.modes.length" accent="violet">
              <div class="space-y-1.5">
                <div v-for="m in arch.grounding.agent.modes" :key="m.name" class="flex items-baseline gap-2">
                  <code class="text-[11px] font-mono font-semibold text-violet-700 flex-shrink-0">{{ m.name }}</code>
                  <span class="text-[11px] text-gray-500 leading-snug">{{ m.desc }}</span>
                </div>
              </div>
            </PeekPanel>
          </div>
        </div>

        <!-- concurrency timing annotation (the engineering highlight) -->
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-rose-200 bg-rose-50/40 px-2.5 py-1.5 mb-2.5">
          <div class="i-lucide-git-compare-arrows text-rose-400 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-rose-700 leading-snug">{{ arch.grounding.agent.concurrency }}</span>
        </div>

        <!-- output -->
        <div class="rounded-xl border border-violet-200 bg-violet-50/40 px-2.5 py-2">
          <div class="flex items-center gap-1.5 mb-1">
            <div class="i-lucide-arrow-down-to-line text-violet-600 text-sm" />
            <span class="text-xs font-bold text-gray-700">OUTPUT · {{ arch.grounding.output.label }}</span>
          </div>
          <div class="flex flex-wrap gap-1">
            <span
              v-for="p in arch.grounding.output.parts"
              :key="p"
              class="px-1.5 py-0.5 rounded-md bg-white border border-violet-200 text-[11px] text-violet-700"
            >{{ p }}</span>
          </div>
        </div>
      </ArchBox>
    </div>
    <!-- right: linking demo -->
    <div>
      <LinkingDemo />
    </div>

    <!-- ════ Stage 3: SQL Generation ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="amber" :items="arch.insights.sqlgen" />
    </div>
    <div>
      <Connector label="GroundedContext → SQL" />
      <ArchBox icon="i-lucide-code-2" :title="arch.sqlgen.title" :role="arch.sqlgen.role" accent="amber">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-2.5 mb-2.5">
          <!-- Prompt -->
          <div class="rounded-xl border p-2.5" :class="ACCENTS.amber.surface">
            <div class="flex items-center gap-1.5 mb-2">
              <div class="i-lucide-square-terminal text-amber-600 text-sm" />
              <span class="text-xs font-bold text-gray-700">Prompt</span>
              <span class="text-[10px] text-gray-400 font-mono ml-auto truncate">{{ arch.sqlgen.prompt.engine }}</span>
            </div>
            <div class="flex flex-wrap gap-1 mb-2">
              <span
                v-for="b in arch.sqlgen.prompt.blocks"
                :key="b.label"
                class="px-1.5 py-0.5 rounded-md bg-white border border-amber-200 text-[11px] font-medium text-amber-700"
                :title="b.desc"
              >{{ b.label }}</span>
            </div>
            <PeekPanel label="SQL 最佳实践" icon="i-lucide-list-checks" :count="arch.sqlgen.prompt.rules.length" accent="amber">
              <ol class="space-y-1.5 pl-0 list-none">
                <li v-for="(r, i) in arch.sqlgen.prompt.rules" :key="i" class="flex items-start gap-2 text-[11px] text-gray-700 leading-relaxed">
                  <span class="w-3.5 h-3.5 rounded-full bg-amber-100 text-amber-700 flex-center text-[8px] font-bold flex-shrink-0 mt-0.5">{{ i + 1 }}</span>
                  <span>{{ r }}</span>
                </li>
              </ol>
            </PeekPanel>
          </div>

          <!-- Tools -->
          <div class="rounded-xl border p-2.5" :class="ACCENTS.blue.surface">
            <div class="flex items-center gap-1.5 mb-2">
              <div class="i-lucide-wrench text-blue-600 text-sm" />
              <span class="text-xs font-bold text-gray-700">Tools</span>
            </div>
            <div class="space-y-1.5">
              <div v-for="t in arch.sqlgen.tools" :key="t.name" class="rounded-lg bg-white border border-blue-100 px-2 py-1.5">
                <code class="text-[11px] font-mono font-bold text-blue-700">{{ t.name }}</code>
                <div class="text-[11px] text-gray-500 leading-snug">{{ t.desc }}</div>
              </div>
            </div>
          </div>
        </div>

        <div class="flex items-center justify-center mb-2.5">
          <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-amber-50 border border-amber-200 text-[11px] font-semibold text-amber-700">
            <div class="i-lucide-repeat text-xs" />
            {{ arch.sqlgen.loop }}
          </div>
        </div>

        <!-- verify gate -->
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-emerald-200 bg-emerald-50/40 px-2.5 py-1.5 mb-2.5">
          <div class="i-lucide-shield-check text-emerald-500 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-emerald-700 leading-snug">{{ arch.sqlgen.verify }}</span>
        </div>

        <div class="rounded-xl border border-amber-200 bg-amber-50/40 px-2.5 py-2 flex items-center gap-1.5">
          <div class="i-lucide-file-check-2 text-amber-600 text-sm" />
          <span class="text-xs font-bold text-gray-700">OUTPUT · {{ arch.sqlgen.output.label }}</span>
          <span class="text-[11px] text-gray-400 ml-auto truncate">{{ arch.sqlgen.output.note }}</span>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />

    <!-- ════ Stage 4: Execute ════ -->
    <div v-if="showNotes" class="hidden lg:block">
      <InsightNotes accent="emerald" :items="arch.insights.execute" />
    </div>
    <div>
      <Connector :label="arch.sqlgen.output.label" />
      <ArchBox icon="i-lucide-play" :title="arch.execute.title" :role="arch.execute.role" accent="emerald">
        <ul class="space-y-1 mb-2">
          <li v-for="(s, i) in arch.execute.steps" :key="i" class="flex items-start gap-2 text-xs text-gray-600 leading-relaxed">
            <div class="i-lucide-check mt-0.5 flex-shrink-0 text-emerald-500" />
            <span>{{ s }}</span>
          </li>
        </ul>
        <div class="rounded-xl border border-emerald-200 bg-emerald-50/40 px-2.5 py-2 flex items-center gap-1.5">
          <div class="i-lucide-table text-emerald-600 text-sm" />
          <span class="text-xs font-bold text-gray-700">{{ arch.execute.output }}</span>
        </div>
      </ArchBox>
    </div>
    <div class="hidden lg:block" />

    <!-- ════ reads · Lakebase (side dependency) ════ -->
    <div class="mt-1" :class="showNotes ? 'lg:col-span-3' : 'lg:col-span-2'">
      <div class="rounded-xl border border-dashed border-indigo-200 bg-indigo-50/30 px-3 py-2">
        <div class="flex items-center gap-1.5 mb-1.5">
          <div class="i-lucide-database text-indigo-500 text-sm" />
          <span class="text-xs font-bold text-indigo-700">{{ arch.reads.label }}</span>
        </div>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
          <div v-for="r in arch.reads.items" :key="r.table" class="flex items-center gap-2">
            <code class="px-2 py-0.5 rounded-md bg-gray-900 text-gray-100 font-mono text-[11px] flex-shrink-0">{{ r.table }}</code>
            <span class="text-[11px] text-gray-500 leading-snug">{{ r.use }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
