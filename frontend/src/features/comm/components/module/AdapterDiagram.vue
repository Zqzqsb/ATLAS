<script setup lang="ts">
import { computed } from 'vue'
import type { AdapterDiagram } from '../../model/comm'

const props = defineProps<{ diagram: AdapterDiagram }>()

const KIND_META: Record<
  'rename' | 'expose' | 'computed' | 'relation',
  { label: string; color: string; bg: string; border: string; dot: string }
> = {
  rename: {
    label: '重命名',
    color: 'text-violet-700',
    bg: 'bg-violet-50',
    border: 'border-violet-300',
    dot: 'stroke-violet-500',
  },
  expose: {
    label: '直接暴露',
    color: 'text-emerald-700',
    bg: 'bg-emerald-50',
    border: 'border-emerald-300',
    dot: 'stroke-emerald-500',
  },
  computed: {
    label: '计算列',
    color: 'text-amber-700',
    bg: 'bg-amber-50',
    border: 'border-amber-300',
    dot: 'stroke-amber-500',
  },
  relation: {
    label: '关系列',
    color: 'text-rose-700',
    bg: 'bg-rose-50',
    border: 'border-rose-300',
    dot: 'stroke-rose-500',
  },
}

/** Y-coordinate (in svg local units) for each side, by index. */
const ROW_H = 36
const TOP_PAD = 18
const physY = (i: number) => TOP_PAD + i * ROW_H + ROW_H / 2
const logY = (i: number) => TOP_PAD + i * ROW_H + ROW_H / 2

const physCount = computed(() => props.diagram.physical.columns.length)
const logCount = computed(() => props.diagram.logical.columns.length)
const svgH = computed(() => TOP_PAD * 2 + Math.max(physCount.value, logCount.value) * ROW_H)

interface MappingLine {
  pIdx: number
  lIdx: number
  kind: 'rename' | 'expose' | 'computed' | 'relation'
}

const lines = computed<MappingLine[]>(() => {
  const out: MappingLine[] = []
  props.diagram.logical.columns.forEach((lc, lIdx) => {
    if (!lc.from) return
    const sources = Array.isArray(lc.from) ? lc.from : [lc.from]
    sources.forEach((src) => {
      const pIdx = props.diagram.physical.columns.findIndex(
        (pc) => pc.name === src && !pc.hidden,
      )
      if (pIdx >= 0) out.push({ pIdx, lIdx, kind: lc.kind })
    })
  })
  return out
})

function pathFor(line: MappingLine): string {
  const y1 = physY(line.pIdx)
  const y2 = logY(line.lIdx)
  // Simple cubic bezier from (0, y1) to (100, y2) in viewBox coords
  return `M 0 ${y1} C 50 ${y1}, 50 ${y2}, 100 ${y2}`
}

function lineClass(kind: MappingLine['kind']) {
  return KIND_META[kind].dot
}
</script>

<template>
  <figure class="rounded-xl border border-gray-200 bg-white p-3 mb-3">
    <figcaption
      v-if="diagram.caption"
      class="text-[10.5px] font-bold tracking-wider text-gray-500 mb-2 flex items-center gap-1.5"
    >
      <div class="i-lucide-git-compare-arrows text-amber-500 text-[12px]" />
      {{ diagram.caption }}
    </figcaption>

    <div class="grid grid-cols-[1fr_120px_1fr] gap-2 items-stretch">
      <!-- ═══ LEFT: physical schema ═══ -->
      <div class="rounded-lg border border-gray-200 bg-gray-50/60 overflow-hidden">
        <div class="px-2.5 py-1.5 bg-gray-100/80 border-b border-gray-200 flex items-baseline gap-1.5">
          <div class="i-lucide-database text-gray-500 text-[12px] flex-shrink-0" />
          <span class="text-[11px] font-bold text-gray-700">{{ diagram.physical.label }}</span>
          <span v-if="diagram.physical.sublabel" class="text-[10px] font-mono text-gray-400 ml-auto">
            {{ diagram.physical.sublabel }}
          </span>
        </div>
        <div class="py-1">
          <div
            v-for="(col, i) in diagram.physical.columns"
            :key="col.name + '__' + i"
            class="px-2.5 flex items-center gap-1.5 text-[11px] font-mono"
            :style="{ height: '36px' }"
            :class="col.hidden ? 'text-gray-400 line-through opacity-60' : 'text-gray-800'"
          >
            <span class="flex-1">{{ col.name }}</span>
            <span v-if="col.type" class="text-[9.5px] text-gray-400 font-normal">{{ col.type }}</span>
            <span
              v-if="col.sensitive"
              class="text-[8.5px] font-bold tracking-wider text-rose-700 bg-rose-100 px-1 rounded leading-none"
            >敏感 · 隐藏</span>
            <span
              v-else-if="col.hidden"
              class="text-[8.5px] font-bold tracking-wider text-gray-500 bg-gray-200 px-1 rounded leading-none"
            >隐藏</span>
          </div>
        </div>
      </div>

      <!-- ═══ MIDDLE: mapping lines (svg) ═══ -->
      <div class="relative">
        <svg
          :viewBox="`0 0 100 ${svgH}`"
          :height="svgH"
          width="100%"
          preserveAspectRatio="none"
          class="overflow-visible"
        >
          <path
            v-for="(line, i) in lines"
            :key="i"
            :d="pathFor(line)"
            fill="none"
            stroke="currentColor"
            stroke-width="1.4"
            stroke-linecap="round"
            :class="lineClass(line.kind)"
          />
        </svg>
        <!-- adapter label centered -->
        <div class="absolute inset-0 flex items-center justify-center pointer-events-none">
          <div class="bg-white/90 border border-amber-300 rounded-md px-2 py-0.5 shadow-sm flex items-center gap-1">
            <div class="i-lucide-shuffle text-amber-600 text-[11px]" />
            <span class="text-[10px] font-bold text-amber-700">Adapter</span>
          </div>
        </div>
      </div>

      <!-- ═══ RIGHT: logical model ═══ -->
      <div class="rounded-lg border border-amber-200 bg-amber-50/30 overflow-hidden">
        <div class="px-2.5 py-1.5 bg-amber-100/60 border-b border-amber-200 flex items-baseline gap-1.5">
          <div class="i-lucide-layers text-amber-600 text-[12px] flex-shrink-0" />
          <span class="text-[11px] font-bold text-amber-800">{{ diagram.logical.label }}</span>
          <span v-if="diagram.logical.sublabel" class="text-[10px] font-mono text-amber-600/70 ml-auto">
            {{ diagram.logical.sublabel }}
          </span>
        </div>
        <div class="py-1">
          <div
            v-for="(col, i) in diagram.logical.columns"
            :key="col.name + '__' + i"
            class="px-2.5 flex flex-col justify-center"
            :style="{ height: '36px' }"
          >
            <div class="flex items-center gap-1.5 text-[11px] font-mono">
              <span class="text-gray-900 font-semibold">{{ col.name }}</span>
              <span
                class="text-[8.5px] font-bold tracking-wider px-1 rounded leading-none"
                :class="[KIND_META[col.kind].color, KIND_META[col.kind].bg, 'border', KIND_META[col.kind].border]"
              >{{ KIND_META[col.kind].label }}</span>
              <span v-if="col.expr" class="text-[9.5px] text-sky-700 bg-sky-50 border border-sky-200/70 font-mono px-1 py-0.5 rounded leading-none truncate">
                = {{ col.expr }}
              </span>
            </div>
            <p v-if="col.note" class="text-[9.5px] text-gray-500 leading-tight mt-0.5">{{ col.note }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- legend -->
    <div class="flex items-center gap-3 mt-2 px-1 flex-wrap">
      <span
        v-for="(meta, key) in KIND_META"
        :key="key"
        class="inline-flex items-center gap-1 text-[9.5px]"
      >
        <svg width="14" height="6" class="overflow-visible">
          <path d="M 0 3 L 14 3" stroke="currentColor" stroke-width="1.4" :class="meta.dot" fill="none" />
        </svg>
        <span class="font-semibold" :class="meta.color">{{ meta.label }}</span>
      </span>
      <span class="inline-flex items-center gap-1 text-[9.5px] ml-2">
        <span class="text-gray-400 line-through font-mono">ssn</span>
        <span class="text-gray-500">不映射 → 完全隐藏</span>
      </span>
    </div>
  </figure>
</template>
