<script setup lang="ts">
/**
 * VendorStage — full-width vendor showcase with two-level switching.
 *
 *   ┌─ axis tabs ───────────────────────────────────────────────┐
 *   │  [⚪ 白盒 · 开源 · 4]   [⚫ 黑盒 · 托管 · 2]                 │
 *   ├─ vendor pills (within active axis) ───────────────────────┤
 *   │  [主 WrenAI] [ATLAS] [dbt SL] [Cube]                       │
 *   ├─ one full-width vendor card (detail + example always shown)
 *   │  ...                                                       │
 *   └────────────────────────────────────────────────────────────┘
 */
import { computed, ref, watch } from 'vue'
import type { VendorTake } from '../../model/comm'
import { ACCENTS, SCHOOL_META, splitTakesByAxis } from '../../model/comm'
import { COMM_SOURCES } from '../../model/comm-sources'
import CodeRefChip from './CodeRefChip.vue'
import EvidenceChip from '../../../arch/components/module/diagram/EvidenceChip.vue'
import AdapterDiagram from './AdapterDiagram.vue'
import DetailPanel from './DetailPanel.vue'
import CodeBlock from './CodeBlock.vue'
import InlineCode from './InlineCode.vue'

const props = defineProps<{
  takes: VendorTake[]
  scopeKey: string
}>()

const split = computed(() => splitTakesByAxis(props.takes))

function sorted(list: VendorTake[]) {
  const out = [...list]
  out.sort((a, b) => Number(b.primary ?? false) - Number(a.primary ?? false))
  return out
}
const whites = computed(() => sorted(split.value.white))
const blacks = computed(() => sorted(split.value.black))

const axis = ref<'white' | 'black'>('white')
const vendorIdx = ref(0)

/* ─── per-card collapse state ──────────────────────────────────────
 * Keyed by `${axis}@${vendor}@${idx}` so each card remembers its own
 * open/closed state independently. Default = collapsed (frame-first view). */
const diagramOpen = ref<Record<string, boolean>>({})
const exampleOpen = ref<Record<string, boolean>>({})

watch(
  () => props.scopeKey,
  () => {
    axis.value = whites.value.length ? 'white' : 'black'
    vendorIdx.value = 0
    diagramOpen.value = {}
    exampleOpen.value = {}
  },
  { immediate: true },
)

const list = computed(() => (axis.value === 'white' ? whites.value : blacks.value))
const active = computed(() => list.value[vendorIdx.value])
const cardKey = computed(() => `${axis.value}@${active.value?.vendor ?? ''}@${vendorIdx.value}`)

function setAxis(a: 'white' | 'black') {
  if (axis.value === a) return
  axis.value = a
  vendorIdx.value = 0
}
function pick(i: number) {
  vendorIdx.value = i
}
function toggleDiagram() {
  diagramOpen.value[cardKey.value] = !diagramOpen.value[cardKey.value]
}
function toggleExample() {
  exampleOpen.value[cardKey.value] = !exampleOpen.value[cardKey.value]
}
</script>

<template>
  <div class="rounded-2xl border border-gray-200 bg-white overflow-hidden">
    <!-- ════ axis tabs ════ -->
    <div class="px-3 pt-3 pb-0 flex items-center gap-2 border-b border-gray-100 bg-gray-50/50">
      <button
        v-if="whites.length"
        class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-t-lg text-[11.5px] font-bold transition-all border border-b-0"
        :class="[
          axis === 'white'
            ? 'bg-white border-emerald-200 text-emerald-700 -mb-px relative z-10'
            : 'bg-transparent border-transparent text-gray-500 hover:text-gray-700',
        ]"
        @click="setAxis('white')"
      >
        <span class="w-1.5 h-1.5 rounded-full" :class="ACCENTS.emerald.dot" />
        白盒 · 开源
        <span class="text-[10px] font-mono opacity-70">{{ whites.length }}</span>
      </button>
      <button
        v-if="blacks.length"
        class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-t-lg text-[11.5px] font-bold transition-all border border-b-0"
        :class="[
          axis === 'black'
            ? 'bg-white border-blue-200 text-blue-700 -mb-px relative z-10'
            : 'bg-transparent border-transparent text-gray-500 hover:text-gray-700',
        ]"
        @click="setAxis('black')"
      >
        <span class="w-1.5 h-1.5 rounded-full" :class="ACCENTS.blue.dot" />
        黑盒 · 托管
        <span class="text-[10px] font-mono opacity-70">{{ blacks.length }}</span>
      </button>
    </div>

    <!-- ════ vendor pills ════ -->
    <div
      v-if="list.length"
      class="px-3 py-2 flex items-center gap-1 flex-wrap border-b border-gray-100 bg-white"
    >
      <button
        v-for="(t, i) in list"
        :key="t.vendor + '__' + i"
        class="inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-[11.5px] font-semibold border transition-all"
        :class="[
          i === vendorIdx
            ? (axis === 'white'
              ? 'bg-emerald-50 border-emerald-300 text-emerald-700 shadow-sm'
              : 'bg-blue-50 border-blue-300 text-blue-700 shadow-sm')
            : 'bg-white border-gray-200 text-gray-500 hover:border-gray-300 hover:text-gray-700',
        ]"
        @click="pick(i)"
      >
        <span
          v-if="t.primary"
          class="text-[8.5px] font-bold tracking-wider text-amber-700 bg-amber-100 px-1 rounded leading-none"
        >主</span>
        <span>{{ t.vendor }}</span>
      </button>
    </div>

    <!-- ════ vendor card body ════ -->
    <div v-if="active" :key="axis + '@' + active.vendor + '@' + vendorIdx" class="p-4">
      <!-- header line -->
      <div class="flex items-baseline gap-2 mb-2 flex-wrap">
        <div
          class="w-2 h-2 rounded-full flex-shrink-0"
          :class="ACCENTS[SCHOOL_META[active.school].accent].dot"
        />
        <h4 class="text-[16px] font-extrabold text-gray-900 leading-tight">{{ active.vendor }}</h4>
        <span
          v-if="active.primary"
          class="text-[9px] font-bold tracking-wider text-amber-700 bg-amber-100 px-1.5 py-0.5 rounded leading-none"
        >主轴</span>
        <span
          class="text-[10.5px] font-mono"
          :class="ACCENTS[SCHOOL_META[active.school].accent].text"
        >{{ SCHOOL_META[active.school].label }}</span>
      </div>

      <!-- one-liner -->
      <p class="text-[13px] text-gray-700 leading-relaxed mb-2.5"><InlineCode :text="active.desc" /></p>

      <!-- ref chips -->
      <div
        v-if="active.code?.length || active.refs?.length"
        class="flex flex-wrap items-center gap-1.5 mb-3"
      >
        <CodeRefChip v-if="active.code?.length" :refs="active.code" size="sm" />
        <EvidenceChip
          v-if="active.refs?.length"
          :refs="active.refs"
          :catalog="COMM_SOURCES"
          size="sm"
        />
      </div>

      <!-- detail prose / structured -->
      <div v-if="active.detail" class="mb-2.5">
        <DetailPanel
          v-if="typeof active.detail === 'object'"
          :detail="active.detail"
          :accent="SCHOOL_META[active.school].accent"
          :default-open="false"
        />
        <div
          v-else
          class="rounded-lg border bg-gray-50/60 px-3.5 py-3"
          :class="`border-${SCHOOL_META[active.school].accent}-200/70`"
        >
          <div class="flex items-center gap-1.5 mb-2">
            <div class="i-lucide-book-open text-[12px]" :class="ACCENTS[SCHOOL_META[active.school].accent].text" />
            <span class="text-[10px] font-bold tracking-wider text-gray-500">DETAIL · 怎么做</span>
          </div>
          <p
            class="text-[12px] text-gray-700 leading-relaxed whitespace-pre-line"
          ><InlineCode :text="String(active.detail)" /></p>
        </div>
      </div>

      <!-- adapter diagram (collapsible · default closed to keep the frame view tight) -->
      <div
        v-if="active.diagram"
        class="rounded-xl border border-gray-200 bg-white overflow-hidden mb-2.5"
      >
        <button
          type="button"
          class="w-full flex items-center gap-2 px-3 py-2 text-left transition-colors hover:bg-gray-50"
          :class="diagramOpen[cardKey] ? 'border-b border-gray-100' : ''"
          @click="toggleDiagram"
        >
          <div
            class="w-5 h-5 rounded-md flex-center flex-shrink-0"
            :class="[ACCENTS[SCHOOL_META[active.school].accent].iconBg, ACCENTS[SCHOOL_META[active.school].accent].iconText]"
          >
            <div class="i-lucide-arrow-left-right text-[11px]" />
          </div>
          <span class="text-[10px] font-bold tracking-wider text-gray-500">ADAPTER · 物理 ↔ 逻辑</span>
          <span
            v-if="active.diagram.caption"
            class="text-[12px] text-gray-700 truncate"
          >{{ active.diagram.caption }}</span>
          <span class="ml-auto text-[10px] text-gray-400">
            {{ diagramOpen[cardKey] ? '点击折叠' : '点击展开示意图' }}
          </span>
          <div
            class="i-lucide-chevron-down text-gray-400 text-sm flex-shrink-0 transition-transform"
            :class="{ 'rotate-180': diagramOpen[cardKey] }"
          />
        </button>
        <Transition
          enter-active-class="transition-all duration-200 ease-out overflow-hidden"
          leave-active-class="transition-all duration-150 ease-in overflow-hidden"
          enter-from-class="opacity-0 max-h-0"
          enter-to-class="opacity-100 max-h-[1600px]"
          leave-from-class="opacity-100 max-h-[1600px]"
          leave-to-class="opacity-0 max-h-0"
        >
          <div v-show="diagramOpen[cardKey]" class="p-3">
            <AdapterDiagram :diagram="active.diagram" />
          </div>
        </Transition>
      </div>

      <!-- example code (collapsible · default closed) -->
      <div
        v-if="active.example"
        class="rounded-lg border border-gray-800 overflow-hidden bg-gray-900"
      >
        <button
          type="button"
          class="w-full flex items-center gap-2 px-3 py-1.5 text-left transition-colors hover:bg-gray-700/40 bg-gray-800"
          @click="toggleExample"
        >
          <span
            v-if="active.example.lang"
            class="px-1.5 py-0.5 rounded bg-gray-700 text-gray-200 text-[10.5px] font-mono"
          >{{ active.example.lang }}</span>
          <span
            v-if="active.example.caption"
            class="text-[10.5px] font-mono text-gray-400 truncate"
          >{{ active.example.caption }}</span>
          <span class="ml-auto text-[9.5px] text-gray-500 tracking-wider">
            EXAMPLE · {{ exampleOpen[cardKey] ? '点击折叠' : '点击展开代码' }}
          </span>
          <div
            class="i-lucide-chevron-down text-gray-500 text-sm flex-shrink-0 transition-transform"
            :class="{ 'rotate-180': exampleOpen[cardKey] }"
          />
        </button>
        <Transition
          enter-active-class="transition-all duration-200 ease-out overflow-hidden"
          leave-active-class="transition-all duration-150 ease-in overflow-hidden"
          enter-from-class="opacity-0 max-h-0"
          enter-to-class="opacity-100 max-h-[1600px]"
          leave-from-class="opacity-100 max-h-[1600px]"
          leave-to-class="opacity-0 max-h-0"
        >
          <div v-show="exampleOpen[cardKey]" class="border-t border-gray-700">
            <CodeBlock :code="active.example.code" :lang="active.example.lang" />
          </div>
        </Transition>
      </div>

      <div
        v-if="!active.detail && !active.example"
        class="text-[11.5px] text-gray-400 italic px-1"
      >
        （这家在本步骤的细节尚未展开）
      </div>
    </div>
  </div>
</template>
