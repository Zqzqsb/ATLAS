<script setup lang="ts">
import { computed } from 'vue'
import { ACCENTS } from '../../../model/architecture'
import type { FlowDef } from '../../../model/flows'
import { getModule } from '../../../model/modules'
import ArchBox from '../diagram/ArchBox.vue'
import Connector from '../diagram/Connector.vue'
import PeekPanel from '../diagram/PeekPanel.vue'
import ChunkTreemap from '../diagram/ChunkTreemap.vue'

const props = defineProps<{ flow: FlowDef }>()
const arch = computed(() => getModule(props.flow.id)?.onboarding ?? null)
</script>

<template>
  <div v-if="arch" class="grid grid-cols-1 lg:grid-cols-[minmax(0,1.05fr)_minmax(0,1fr)] gap-x-7 gap-y-3 items-start">
    <!-- ════ Stage 1: Input ════ -->
    <ArchBox icon="i-lucide-table-2" :title="arch.input.label" accent="slate" muted>
      <div class="flex items-center gap-2 text-xs text-gray-500">
        <code class="px-1.5 py-0.5 rounded bg-gray-100 text-gray-600 font-mono text-[11px]">{{ arch.input.table }}</code>
        <span>{{ arch.input.note }}</span>
      </div>
    </ArchBox>
    <!-- right: input annotation -->
    <div class="hidden lg:flex items-center gap-2 text-xs text-gray-400 px-1 pt-3">
      <div class="i-lucide-corner-left-down text-gray-300" />
      {{ arch.insights.input }}
    </div>

    <!-- ════ Stage 2: Coordinator + Worker (merged) ════ -->
    <div>
      <Connector />
      <ArchBox icon="i-lucide-split" :title="arch.coordinator.title" :role="arch.coordinator.role" accent="violet">
        <ul class="space-y-1 mb-2">
          <li v-for="(p, i) in arch.coordinator.points" :key="i" class="flex items-start gap-2 text-xs text-gray-600 leading-relaxed">
            <div class="i-lucide-check mt-0.5 flex-shrink-0 text-violet-500" />
            <span>{{ p }}</span>
          </li>
        </ul>
        <div class="flex items-start gap-2 rounded-lg border border-dashed border-violet-200 bg-violet-50/40 px-2.5 py-1.5">
          <div class="i-lucide-info text-violet-400 text-xs mt-0.5 flex-shrink-0" />
          <span class="text-[11px] text-violet-700 leading-snug">{{ arch.coordinator.note }}</span>
        </div>
      </ArchBox>

      <Connector :label="arch.worker.dispatch" />

      <ArchBox icon="i-lucide-bot" :title="arch.worker.title" :role="arch.worker.role" accent="emerald" badge="× N">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-2.5 mb-2.5">
          <!-- Prompt -->
          <div class="rounded-xl border p-2.5" :class="ACCENTS.amber.surface">
            <div class="flex items-center gap-1.5 mb-2">
              <div class="i-lucide-square-terminal text-amber-600 text-sm" />
              <span class="text-xs font-bold text-gray-700">Prompt</span>
              <span class="text-[10px] text-gray-400 font-mono ml-auto truncate">{{ arch.worker.prompt.engine }}</span>
            </div>
            <div class="flex flex-wrap gap-1">
              <span
                v-for="b in arch.worker.prompt.blocks"
                :key="b.label"
                class="px-1.5 py-0.5 rounded-md bg-white border border-amber-200 text-[11px] font-medium text-amber-700"
                :title="b.desc"
              >{{ b.label }}</span>
            </div>
          </div>

          <!-- Tools -->
          <div class="rounded-xl border p-2.5" :class="ACCENTS.blue.surface">
            <div class="flex items-center gap-1.5 mb-2">
              <div class="i-lucide-wrench text-blue-600 text-sm" />
              <span class="text-xs font-bold text-gray-700">Tools</span>
            </div>
            <div class="space-y-1.5">
              <div v-for="t in arch.worker.tools" :key="t.name" class="rounded-lg bg-white border border-blue-100 px-2 py-1.5">
                <code class="text-[11px] font-mono font-bold text-blue-700">{{ t.name }}</code>
                <div class="text-[11px] text-gray-500 leading-snug">{{ t.desc }}</div>
              </div>
            </div>
          </div>
        </div>

        <div class="flex items-center justify-center mb-2.5">
          <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-emerald-50 border border-emerald-200 text-[11px] font-semibold text-emerald-700">
            <div class="i-lucide-repeat text-xs" />
            {{ arch.worker.loop }}
          </div>
        </div>

        <div class="rounded-xl border border-emerald-200 bg-emerald-50/40 px-2.5 py-2 flex items-center gap-1.5">
          <div class="i-lucide-arrow-down-to-line text-emerald-600 text-sm" />
          <span class="text-xs font-bold text-gray-700">OUTPUT · {{ arch.worker.output.label }}</span>
          <span class="text-[11px] text-gray-400">{{ arch.worker.output.types.length }} 类</span>
          <code class="text-[10px] font-mono text-emerald-700 bg-white border border-emerald-200 rounded px-1 ml-auto">{{ arch.worker.output.store }}</code>
        </div>
      </ArchBox>
    </div>

    <!-- right: treemap demo + worker reference details + process insights -->
    <div class="space-y-3 lg:pt-7">
      <ChunkTreemap />

      <!-- Worker reference details (moved out of the spine to keep it structural) -->
      <div class="space-y-1.5">
        <div class="flex items-center gap-1.5 text-[11px] font-semibold text-gray-400 uppercase tracking-wide">
          <div class="i-lucide-bot text-gray-300" /> Worker 细节
        </div>
        <PeekPanel label="关键约束 / 技巧" icon="i-lucide-shield-alert" :count="arch.worker.prompt.rules.length" accent="amber">
          <ol class="space-y-1.5">
            <li v-for="(r, i) in arch.worker.prompt.rules" :key="i" class="flex items-start gap-2 text-[11px] text-gray-700 leading-relaxed">
              <span class="w-3.5 h-3.5 rounded-full bg-amber-100 text-amber-700 flex-center text-[8px] font-bold flex-shrink-0 mt-0.5">{{ i + 1 }}</span>
              <span>{{ r }}</span>
            </li>
          </ol>
        </PeekPanel>
        <PeekPanel :label="`${arch.worker.output.label} 的 ${arch.worker.output.types.length} 类内容`" icon="i-lucide-tags" :count="arch.worker.output.types.length" accent="emerald">
          <div class="space-y-1.5">
            <div v-for="t in arch.worker.output.types" :key="t.name" class="flex items-baseline gap-2">
              <code class="text-[11px] font-mono font-semibold text-emerald-700 flex-shrink-0">{{ t.name }}</code>
              <span class="text-[11px] text-gray-500 leading-snug">{{ t.desc }}</span>
            </div>
          </div>
        </PeekPanel>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-1 gap-2">
        <div
          v-for="ins in arch.insights.process"
          :key="ins.title"
          class="rounded-xl border border-gray-200 bg-gradient-to-br from-white to-gray-50/60 px-3 py-2"
        >
          <div class="flex items-center gap-1.5 mb-0.5">
            <div class="w-5 h-5 rounded-md bg-emerald-50 flex-center"><div :class="[ins.icon, 'text-emerald-600 text-[11px]']" /></div>
            <span class="text-xs font-bold text-gray-800">{{ ins.title }}</span>
          </div>
          <p class="text-[11px] text-gray-500 leading-relaxed">{{ ins.body }}</p>
        </div>
      </div>
    </div>

    <!-- ════ Stage 3: Storage ════ -->
    <div>
      <Connector label="RC produced" />
      <ArchBox icon="i-lucide-database" :title="arch.storage.title" accent="indigo">
        <div class="space-y-1.5">
          <div v-for="item in arch.storage.items" :key="item.table" class="flex items-center gap-2.5">
            <code class="px-2 py-0.5 rounded-md bg-gray-900 text-gray-100 font-mono text-[11px] flex-shrink-0">{{ item.table }}</code>
            <div class="flex-1 min-w-0">
              <div class="text-xs font-semibold text-gray-800">{{ item.label }}</div>
              <div class="text-[11px] text-gray-400 truncate">{{ item.note }}</div>
            </div>
            <code class="hidden md:block text-[10px] font-mono text-gray-400 flex-shrink-0">{{ item.spec }}</code>
          </div>
        </div>
      </ArchBox>
    </div>

    <!-- right: storage insights -->
    <div class="space-y-2 lg:pt-9">
      <div
        v-for="ins in arch.insights.storage"
        :key="ins.title"
        class="rounded-xl border border-gray-200 bg-gradient-to-br from-white to-gray-50/60 px-3 py-2"
      >
        <div class="flex items-center gap-1.5 mb-0.5">
          <div class="w-5 h-5 rounded-md bg-indigo-50 flex-center"><div :class="[ins.icon, 'text-indigo-600 text-[11px]']" /></div>
          <span class="text-xs font-bold text-gray-800">{{ ins.title }}</span>
        </div>
        <p class="text-[11px] text-gray-500 leading-relaxed">{{ ins.body }}</p>
      </div>
    </div>
  </div>
</template>
