<script setup lang="ts">
import { computed } from 'vue'
import { ACCENTS, SCHOOL_META, getCommStage, type CommFlowDef } from '../../model/comm'
import VendorStage from './VendorStage.vue'

const props = defineProps<{ flow: CommFlowDef; showNotes?: boolean }>()
const arch = computed(() => getCommStage(props.flow.id))
const a = computed(() => ACCENTS[props.flow.accent])

/** Layout when steps[] are present: vendor showcase takes ~80% */
const stepGridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.18fr)_minmax(0,1fr)]'
    : 'lg:grid-cols-1',
)
/** Layout when legacy variants[] are present (kept for older flows) */
const legacyGridCols = computed(() =>
  props.showNotes
    ? 'lg:grid-cols-[minmax(0,0.85fr)_minmax(0,1.4fr)_minmax(0,0.95fr)]'
    : 'lg:grid-cols-[minmax(0,1.2fr)_minmax(0,1fr)]',
)
</script>

<template>
  <div v-if="arch">
    <!-- ════ Top: stage abstract + principles strip ════ -->
    <div class="rounded-2xl border bg-white px-5 py-4 mb-4" :class="a.surface">
      <div class="flex items-start gap-3">
        <div class="i-lucide-quote text-2xl flex-shrink-0" :class="a.text" />
        <p class="text-[13px] text-gray-700 leading-relaxed font-medium">{{ arch.abstract }}</p>
      </div>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2 mb-6">
      <div
        v-for="(p, i) in arch.principles"
        :key="p.name"
        class="rounded-xl border border-gray-200 bg-gradient-to-br from-white to-gray-50/60 px-3 py-2.5"
      >
        <div class="flex items-center gap-1.5 mb-1">
          <div class="w-5 h-5 rounded-md flex-center text-[10px] font-bold" :class="[a.iconBg, a.iconText]">
            {{ i + 1 }}
          </div>
          <span class="text-[12px] font-bold text-gray-800 leading-tight">{{ p.name }}</span>
        </div>
        <p class="text-[11px] text-gray-500 leading-relaxed">{{ p.desc }}</p>
      </div>
    </div>

    <!-- ════ Sub-questions: each is a 3-column row (notes ｜ variants ｜ commonSense) ════ -->
    <div class="space-y-5">
      <div
        v-for="(q, qi) in arch.subQuestions"
        :key="q.id"
        class="rounded-2xl border border-gray-200 bg-white overflow-hidden"
      >
        <!-- header -->
        <div class="px-5 py-3 border-b border-gray-100 bg-gradient-to-r" :class="`from-${flow.accent}-50/50 to-transparent`">
          <div class="flex items-baseline gap-2.5">
            <div class="text-[10px] font-bold tracking-wider px-1.5 py-0.5 rounded" :class="[a.chip]">
              Q{{ qi + 1 }}
            </div>
            <h3 class="text-[15px] font-extrabold text-gray-900 leading-tight">{{ q.question }}</h3>
          </div>
          <p class="text-[11.5px] text-gray-500 mt-1 leading-snug">{{ q.why }}</p>
        </div>

        <!-- body: steps[] use a vertical full-width vendor showcase; legacy variants[] keep 3-col grid -->
        <div
          v-if="q.steps && q.steps.length"
          class="grid grid-cols-1 gap-x-5 gap-y-3 p-4 items-start"
          :class="stepGridCols"
        >
          <div v-if="showNotes" class="hidden lg:block">
            <div class="rounded-xl border border-dashed border-gray-300 bg-gray-50/60 px-3 py-2.5 sticky top-2">
              <div class="flex items-center gap-1.5 mb-1">
                <div class="i-lucide-sticky-note text-gray-400 text-xs" />
                <span class="text-[10px] font-bold text-gray-500 tracking-wider">讲解备注</span>
              </div>
              <p class="text-[11px] text-gray-500 leading-relaxed">
                公约数 {{ q.steps.length }} 步——每一步先看 <b class="text-emerald-700">白盒</b>（WrenAI 主轴）再看 <b class="text-blue-700">黑盒</b>（Cortex 主轴）。点 vendor pill 切产品；点 <code class="font-mono text-violet-600">file</code> 跳 GitHub，<code class="font-mono text-blue-600">[Sn]</code> 跳官方文档。
              </p>
              <p class="text-[10.5px] text-gray-500 leading-relaxed mt-2 pt-2 border-t border-dashed border-gray-200">
                <b class="text-gray-600">DETAIL · ADAPTER · EXAMPLE 三块默认折叠</b>——讲框架时保持页面紧凑，需要展开哪一项再点对应横条。
              </p>
            </div>
          </div>

          <div class="space-y-3">
            <div
              v-for="(step, si) in q.steps"
              :key="step.id"
              class="rounded-xl border border-gray-200 bg-white overflow-hidden"
            >
              <div class="px-3 py-2 border-b border-gray-100 bg-gradient-to-r from-gray-50/80 to-transparent flex items-baseline gap-2">
                <div
                  class="w-5 h-5 rounded-md flex-center text-[10px] font-bold flex-shrink-0"
                  :class="[a.iconBg, a.iconText]"
                >
                  {{ si + 1 }}
                </div>
                <div :class="[step.icon ?? 'i-lucide-circle-dot', a.text, 'text-sm flex-shrink-0']" />
                <span class="text-[12.5px] font-bold text-gray-800">{{ step.name }}</span>
                <span class="text-[10.5px] text-gray-500 leading-snug ml-1">{{ step.desc }}</span>
              </div>

              <div class="p-3">
                <VendorStage
                  :takes="step.takes"
                  :scope-key="q.id + '::' + step.id"
                />
              </div>
            </div>

            <!-- common-sense block stays under the showcase for steps[] mode -->
            <div class="rounded-2xl border border-rose-200 bg-gradient-to-br from-rose-50/70 to-white p-3.5">
              <div class="flex items-center gap-1.5 mb-1.5">
                <div class="i-lucide-lightbulb text-rose-500 text-sm" />
                <span class="text-[11px] font-bold tracking-wider text-rose-600">COMMON SENSE</span>
              </div>
              <p class="text-[12px] text-gray-700 leading-relaxed">{{ q.commonSense }}</p>
            </div>
          </div>
        </div>

        <!-- legacy variants[] body grid -->
        <div
          v-else
          class="grid grid-cols-1 gap-x-5 gap-y-3 p-4 items-start"
          :class="legacyGridCols"
        >
          <div v-if="showNotes" class="hidden lg:block">
            <div class="rounded-xl border border-dashed border-gray-300 bg-gray-50/60 px-3 py-2.5">
              <div class="flex items-center gap-1.5 mb-1">
                <div class="i-lucide-sticky-note text-gray-400 text-xs" />
                <span class="text-[10px] font-bold text-gray-500 tracking-wider">讲解备注</span>
              </div>
              <p v-if="q.variants?.length" class="text-[11px] text-gray-500 leading-relaxed">
                这一问把这个 stage 在不同体系下"怎么实现"细分成 {{ q.variants.length }} 种取舍。每种取舍下面列出有代表性的产品。
              </p>
            </div>
          </div>

          <!-- variants: data-driven body -->
          <div class="space-y-2">
            <template v-if="q.variants">
              <div
                v-for="v in q.variants"
                :key="v.name"
                class="rounded-xl border p-2.5"
                :class="ACCENTS[v.accent].surface"
              >
                <div class="flex items-baseline gap-2 mb-1">
                  <div class="w-1.5 h-1.5 rounded-full" :class="ACCENTS[v.accent].dot" />
                  <span class="text-[12.5px] font-bold text-gray-800">{{ v.name }}</span>
                  <span class="text-[10.5px] text-gray-500 ml-1.5 leading-snug">{{ v.desc }}</span>
                </div>
                <div class="flex flex-wrap gap-1 mt-1.5">
                  <span
                    v-for="ven in v.vendors"
                    :key="ven"
                    class="px-2 py-0.5 rounded-md text-[10.5px] font-mono font-semibold border bg-white"
                    :class="ACCENTS[v.accent].text + ' border-' + v.accent + '-200'"
                  >{{ ven }}</span>
                </div>
              </div>
            </template>
          </div>

          <!-- commonSense: our framework opinion -->
          <div class="rounded-2xl border border-rose-200 bg-gradient-to-br from-rose-50/70 to-white p-3.5">
            <div class="flex items-center gap-1.5 mb-1.5">
              <div class="i-lucide-lightbulb text-rose-500 text-sm" />
              <span class="text-[11px] font-bold tracking-wider text-rose-600">COMMON SENSE</span>
            </div>
            <p class="text-[12px] text-gray-700 leading-relaxed">{{ q.commonSense }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- ════ Bottom: insights + matrix preview ════ -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 mt-6">
      <!-- Insights -->
      <div class="space-y-2">
        <div class="flex items-center gap-1.5 mb-1">
          <div class="i-lucide-zap text-amber-500 text-sm" />
          <span class="text-[11px] font-bold tracking-wider text-gray-600">关键洞察 · 选择背后的理由</span>
        </div>
        <div
          v-for="ins in arch.insights"
          :key="ins.title"
          class="rounded-xl border border-gray-200 bg-gradient-to-br from-white to-gray-50/60 px-3.5 py-2.5"
        >
          <div class="flex items-center gap-2 mb-1">
            <div class="w-6 h-6 rounded-lg flex-center" :class="[a.iconBg]">
              <div :class="[ins.icon, a.iconText, 'text-sm']" />
            </div>
            <span class="text-[12.5px] font-bold text-gray-900">{{ ins.title }}</span>
          </div>
          <p class="text-[11.5px] text-gray-600 leading-relaxed">{{ ins.body }}</p>
        </div>
      </div>

      <!-- Matrix preview -->
      <div v-if="arch.matrix" class="rounded-2xl border border-slate-200 bg-white overflow-hidden">
        <div class="flex items-center gap-2 px-3.5 py-2 border-b border-slate-100 bg-slate-50/50">
          <div class="i-lucide-grid-3x3 text-slate-500 text-sm" />
          <span class="text-[12px] font-bold text-slate-700">产品对照表</span>
          <span class="text-[10px] text-slate-400 ml-auto">本阶段 4 子问题压缩视图</span>
        </div>
        <table class="w-full text-[10.5px]">
          <thead>
            <tr class="bg-slate-50/60 text-slate-600">
              <th class="text-left font-bold px-2.5 py-1.5">Vendor</th>
              <th
                v-for="c in arch.matrix.cols"
                :key="c"
                class="text-left font-mono font-bold px-2.5 py-1.5"
              >{{ c }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in arch.matrix.rows" :key="row.vendor" class="border-t border-slate-100">
              <td class="px-2.5 py-1.5 font-bold whitespace-nowrap" :class="`text-${SCHOOL_META[row.school].accent}-700`">
                {{ row.vendor }}
              </td>
              <td v-for="(cell, i) in row.cells" :key="i" class="px-2.5 py-1.5 font-mono text-gray-600">
                {{ cell }}
              </td>
            </tr>
          </tbody>
        </table>
        <!-- school legend -->
        <div class="px-3.5 py-2 border-t border-slate-100 flex flex-wrap gap-2">
          <span
            v-for="(meta, key) in SCHOOL_META"
            :key="key"
            class="inline-flex items-center gap-1 text-[10px]"
          >
            <span class="w-1.5 h-1.5 rounded-full" :class="ACCENTS[meta.accent].dot" />
            <span class="font-semibold" :class="ACCENTS[meta.accent].text">{{ meta.label }}</span>
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
