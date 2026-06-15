<script setup lang="ts">
/**
 * VendorDeck — single-card view, 80%-width vendor showcase.
 *
 * Layout:
 *   ┌──────────────── axis tabs (white | black) ────────────────┐
 *   │  [WrenAI] [ATLAS] [dbt SL] [Cube]  ← vendor pills (intra-axis)
 *   │  ┌──────────────────────────────────────────────────────┐
 *   │  │  vendor card  (always expanded — detail + example)   │
 *   │  └──────────────────────────────────────────────────────┘
 *   └──────────────────────────────────────────────────────────┘
 */
import { computed, ref, watch } from 'vue'
import type { VendorTake } from '../../model/comm'
import { ACCENTS, SCHOOL_META } from '../../model/comm'
import { COMM_SOURCES } from '../../model/comm-sources'
import CodeRefChip from './CodeRefChip.vue'
import EvidenceChip from '../../../arch/components/module/diagram/EvidenceChip.vue'

const props = defineProps<{
  /** vendor takes for this stage/step (single axis: white OR black) */
  takes: VendorTake[]
  /** white = open-source, black = managed/closed */
  axis: 'white' | 'black'
  /** unique id used to namespace state */
  scopeKey: string
}>()

/** primary vendor goes first */
const ordered = computed(() => {
  const list = [...props.takes]
  list.sort((a, b) => Number(b.primary ?? false) - Number(a.primary ?? false))
  return list
})
const total = computed(() => ordered.value.length)

const activeIdx = ref(0)
watch(
  () => props.scopeKey,
  () => {
    activeIdx.value = 0
  },
)

const active = computed(() => ordered.value[activeIdx.value])

const axisLabel = computed(() =>
  props.axis === 'white' ? '白盒 · 开源' : '黑盒 · 托管',
)
const axisAccent = computed(() => (props.axis === 'white' ? 'emerald' : 'blue'))

function pick(i: number) {
  activeIdx.value = i
}
</script>

<template>
  <div v-if="total" class="rounded-2xl border bg-white overflow-hidden" :class="ACCENTS[axisAccent].surface">
    <!-- ════ axis bar: label + vendor pills ════ -->
    <div
      class="px-3 py-2 border-b flex items-center gap-2 flex-wrap bg-gradient-to-r"
      :class="[
        ACCENTS[axisAccent].surface,
        `border-${axisAccent}-100`,
      ]"
    >
      <div class="flex items-center gap-1.5 mr-1">
        <div class="w-1.5 h-1.5 rounded-full" :class="ACCENTS[axisAccent].dot" />
        <span class="text-[10.5px] font-bold tracking-wider" :class="ACCENTS[axisAccent].text">
          {{ axisLabel }}
        </span>
        <span class="text-[10px] text-gray-400">· {{ total }} 家</span>
      </div>

      <!-- vendor pills: click to switch the showcased card (within this axis) -->
      <div class="flex items-center gap-1 flex-wrap ml-auto">
        <button
          v-for="(t, i) in ordered"
          :key="t.vendor + '__' + i"
          class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-semibold border transition-all"
          :class="[
            i === activeIdx
              ? `bg-white shadow-sm border-${axisAccent}-300 ${ACCENTS[axisAccent].text}`
              : 'bg-white/70 border-gray-200 text-gray-500 hover:border-gray-300 hover:text-gray-700',
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
    </div>

    <!-- ════ showcased vendor card ════ -->
    <div v-if="active" :key="active.vendor + '@' + activeIdx" class="p-4">
      <!-- header line -->
      <div class="flex items-baseline gap-2 mb-2 flex-wrap">
        <div
          class="w-2 h-2 rounded-full flex-shrink-0"
          :class="ACCENTS[SCHOOL_META[active.school].accent].dot"
        />
        <h4 class="text-[15px] font-extrabold text-gray-900 leading-tight">{{ active.vendor }}</h4>
        <span
          v-if="active.primary"
          class="text-[9px] font-bold tracking-wider text-amber-700 bg-amber-100 px-1.5 py-0.5 rounded leading-none"
        >主轴</span>
        <span
          class="text-[10px] font-mono"
          :class="ACCENTS[SCHOOL_META[active.school].accent].text"
        >{{ SCHOOL_META[active.school].label }}</span>
      </div>

      <!-- one-liner -->
      <p class="text-[12.5px] text-gray-700 leading-relaxed mb-2">{{ active.desc }}</p>

      <!-- ref chips row -->
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

      <!-- detail prose -->
      <div
        v-if="active.detail"
        class="rounded-lg border bg-gray-50/60 px-3 py-2.5 mb-2"
        :class="`border-${SCHOOL_META[active.school].accent}-200/70`"
      >
        <div class="flex items-center gap-1.5 mb-1.5">
          <div class="i-lucide-book-open text-[11px]" :class="ACCENTS[SCHOOL_META[active.school].accent].text" />
          <span class="text-[10px] font-bold tracking-wider text-gray-500">DETAIL · 怎么做</span>
        </div>
        <p
          class="text-[11.5px] text-gray-700 leading-relaxed whitespace-pre-line"
        >{{ active.detail }}</p>
      </div>

      <!-- example code -->
      <div v-if="active.example" class="rounded-lg border border-gray-800 overflow-hidden bg-gray-900">
        <div
          v-if="active.example.caption || active.example.lang"
          class="px-2.5 py-1.5 bg-gray-800 text-[10px] font-mono text-gray-300 flex items-center gap-2"
        >
          <span v-if="active.example.lang" class="px-1 py-0.5 rounded bg-gray-700 text-gray-200">{{ active.example.lang }}</span>
          <span v-if="active.example.caption" class="text-gray-400">{{ active.example.caption }}</span>
          <span class="ml-auto text-[9.5px] text-gray-500">EXAMPLE</span>
        </div>
        <pre class="px-3 py-2.5 text-[11px] font-mono text-gray-100 overflow-x-auto leading-relaxed"><code>{{ active.example.code }}</code></pre>
      </div>

      <!-- empty-state hint when no detail/example yet -->
      <div
        v-if="!active.detail && !active.example"
        class="text-[11px] text-gray-400 italic px-1"
      >
        （这家在本步骤的细节尚未展开）
      </div>
    </div>
  </div>
</template>
