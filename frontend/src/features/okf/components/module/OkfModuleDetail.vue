<script setup lang="ts">
import { computed, ref } from 'vue'
import { ACCENTS } from '../../../arch/model/architecture'
import { getOkfModule, type OkfSection } from '../../model/modules'
import { okfFlows, type OkfFlowDef } from '../../model/okf'
import CodeBlock from '../../../comm/components/module/CodeBlock.vue'
import InlineCode from '../../../comm/components/module/InlineCode.vue'

const props = defineProps<{ flow: OkfFlowDef }>()
const emit = defineEmits<{ back: [] }>()

const a = computed(() => ACCENTS[props.flow.accent])
const showNotes = ref(true)
const mod = computed(() => getOkfModule(props.flow.id))

/* sibling pointers — when drilling into a node, the row of siblings
   shows up as small pills at the top, so you can hop to a neighbouring
   module without going back. */
const siblings = computed(() => okfFlows.filter((f) => f.id !== props.flow.id))

/* expand state for "anatomy" rows: which labels are open */
const openAnatomyRows = ref<Record<string, boolean>>({})
function toggleAnatomy(key: string) {
  openAnatomyRows.value[key] = !openAnatomyRows.value[key]
}

/* tree viewer collapse state */
const treeOpen = ref(true)
</script>

<template>
  <div class="mx-auto px-6 py-7 transition-[max-width] duration-300" :class="showNotes ? 'max-w-7xl' : 'max-w-5xl'">
    <!-- header -->
    <div class="flex items-start gap-3 mb-5">
      <button
        class="mt-0.5 w-9 h-9 rounded-lg flex-center text-gray-500 border border-gray-200 bg-white hover:bg-gray-50 hover:text-gray-800 transition-colors flex-shrink-0"
        title="返回全景 (Esc)"
        @click="emit('back')"
      >
        <div class="i-lucide-arrow-left" />
      </button>
      <div class="w-11 h-11 rounded-xl flex-center text-white bg-gradient-to-br flex-shrink-0" :class="a.gradient">
        <div :class="[flow.icon, 'text-xl']" />
      </div>
      <div class="flex-1 min-w-0">
        <h2 class="text-xl font-extrabold text-gray-900 m-0">{{ flow.title }}</h2>
        <p class="text-sm text-gray-500 mt-1 leading-snug">{{ flow.subtitle }}</p>
      </div>
      <button
        class="mt-0.5 inline-flex items-center gap-1.5 px-2.5 h-9 rounded-lg text-xs font-semibold border transition-colors flex-shrink-0"
        :class="showNotes
          ? 'border-violet-300 bg-violet-50 text-violet-700'
          : 'border-gray-200 bg-white text-gray-500 hover:bg-gray-50 hover:text-gray-800'"
        :title="showNotes ? '隐藏讲解备注' : '展开讲解备注'"
        @click="showNotes = !showNotes"
      >
        <div :class="showNotes ? 'i-lucide-panel-left-close' : 'i-lucide-sticky-note'" />
        讲解备注
      </button>
    </div>

    <!-- sibling pills — quick hop to other OKF modules -->
    <div class="flex flex-wrap items-center gap-1.5 mb-6 pl-12">
      <span class="text-[10px] font-bold tracking-wider text-gray-400 uppercase">jump to</span>
      <button
        v-for="sib in siblings"
        :key="sib.id"
        class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-semibold border transition-colors"
        :class="sib.id === flow.id
          ? 'border-gray-200 bg-gray-50 text-gray-700'
          : 'border-gray-200 bg-white text-gray-500 hover:bg-gray-50 hover:text-gray-800'"
        :title="sib.subtitle"
        @click="() => { /* navigation handled by parent via the same flow prop */ }"
      >
        <div :class="[sib.icon, 'text-[11px]']" />
        {{ sib.label }}
      </button>
    </div>

    <div v-if="!mod" class="text-center text-gray-400 py-20 text-sm">
      该子模块的内部细节还在补充中。
    </div>

    <div v-else class="grid gap-6 transition-[grid-template-columns] duration-300"
         :class="showNotes ? 'lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)]' : 'lg:grid-cols-1'">
      <!-- LEFT: presenter notes -->
      <aside v-if="showNotes" class="space-y-3">
        <div class="rounded-xl border p-4 bg-white/70 backdrop-blur-sm" :class="a.surface">
          <div class="flex items-center gap-1.5 mb-2">
            <div :class="[flow.icon, a.text, 'text-sm']" />
            <span class="text-[10px] font-bold tracking-wider uppercase" :class="a.text">Abstract</span>
          </div>
          <p class="text-[12.5px] text-gray-700 leading-relaxed">{{ mod.abstract }}</p>
        </div>

        <div class="rounded-xl border p-4 bg-white/70 backdrop-blur-sm border-gray-200/80">
          <div class="text-[10px] font-bold tracking-wider uppercase text-gray-400 mb-2">Common Sense · 设计原则</div>
          <ul class="space-y-1.5">
            <li v-for="(p, i) in mod.principles" :key="i" class="text-[12px] text-gray-700 leading-relaxed">
              <span class="font-bold text-gray-900">· {{ p.name }} —</span>
              <InlineCode :text="p.desc" />
            </li>
          </ul>
        </div>

        <div v-if="mod.insights.length" class="rounded-xl border p-4 bg-gradient-to-br from-rose-50/60 to-amber-50/40 border-rose-200/60">
          <div class="text-[10px] font-bold tracking-wider uppercase text-rose-700 mb-2">Why we love it · 我们的看法</div>
          <ul class="space-y-2">
            <li v-for="(it, i) in mod.insights" :key="i" class="text-[12px] text-gray-700 leading-relaxed">
              <span class="font-bold text-rose-700">★ {{ it.title }} —</span>
              <InlineCode :text="it.body" />
            </li>
          </ul>
        </div>
      </aside>

      <!-- RIGHT: detail sections (kind-dispatched) -->
      <main class="space-y-4">
        <div
          v-for="(sec, sIdx) in mod.sections"
          :key="sIdx"
          class="rounded-xl border bg-white/80 backdrop-blur-sm overflow-hidden"
          :class="a.surface"
        >
          <div class="px-4 py-2.5 flex items-center gap-2 border-b" :class="a.surface">
            <div :class="[a.iconBg, a.iconText, 'w-6 h-6 rounded-md flex-center']">
              <div :class="secKindIcon(sec)" />
            </div>
            <span class="text-[13px] font-extrabold text-gray-900">{{ sec.title }}</span>
          </div>

          <div class="p-4">
            <!-- anatomy: expandable label/detail rows -->
            <div v-if="sec.kind === 'anatomy'" class="space-y-1">
              <button
                v-for="(r, i) in sec.rows"
                :key="i"
                type="button"
                class="w-full text-left rounded-lg border border-gray-200/80 px-3 py-2 hover:bg-gray-50 transition-colors"
                :class="openAnatomyRows[`${sec.title}-${i}`] ? 'bg-gray-50' : 'bg-white'"
                @click="toggleAnatomy(`${sec.title}-${i}`)"
              >
                <div class="flex items-center gap-2">
                  <code class="font-mono text-[12px] font-bold" :class="a.text">{{ r.label }}</code>
                  <div
                    class="i-lucide-chevron-right text-[12px] text-gray-400 ml-auto transition-transform"
                    :class="openAnatomyRows[`${sec.title}-${i}`] ? 'rotate-90' : ''"
                  />
                </div>
                <Transition
                  enter-active-class="transition-all duration-200 ease-out"
                  enter-from-class="opacity-0 -translate-y-1"
                  enter-to-class="opacity-100 translate-y-0"
                  leave-active-class="transition-all duration-150 ease-in"
                  leave-from-class="opacity-100"
                  leave-to-class="opacity-0"
                >
                  <p v-if="openAnatomyRows[`${sec.title}-${i}`]" class="text-[11.5px] text-gray-600 leading-relaxed mt-1.5 pl-1 border-l-2" :class="a.dot + ' border-current/20'">
                    <InlineCode :text="r.detail" />
                  </p>
                </Transition>
              </button>
            </div>

            <!-- loop / flow: ordered step list -->
            <ol v-else-if="sec.kind === 'loop' || sec.kind === 'flow'" class="space-y-1.5">
              <li v-for="(stp, i) in sec.steps" :key="i" class="flex items-start gap-2 rounded-lg border border-gray-200/80 bg-white px-3 py-2">
                <div class="w-5 h-5 rounded-md flex-center flex-shrink-0 text-[11px] font-bold text-white" :class="a.gradient">
                  {{ i + 1 }}
                </div>
                <div class="min-w-0 flex-1">
                  <code class="font-mono text-[12px] font-bold text-gray-900">{{ stp.name }}</code>
                  <p class="text-[11.5px] text-gray-600 leading-relaxed mt-0.5">
                    <InlineCode :text="stp.desc" />
                  </p>
                </div>
              </li>
            </ol>

            <!-- tree: monospace ascii block -->
            <div v-else-if="sec.kind === 'tree'">
              <button
                type="button"
                class="text-[10px] font-bold tracking-wider text-gray-500 hover:text-gray-800 transition-colors flex items-center gap-1 mb-1"
                @click="treeOpen = !treeOpen"
              >
                <div class="i-lucide-chevron-right text-[10px] transition-transform" :class="treeOpen ? 'rotate-90' : ''" />
                {{ treeOpen ? 'collapse' : 'expand' }}
              </button>
              <Transition
                enter-active-class="transition-all duration-200 ease-out"
                enter-from-class="opacity-0 max-h-0"
                enter-to-class="opacity-100 max-h-[600px]"
                leave-active-class="transition-all duration-150 ease-in"
                leave-from-class="opacity-100 max-h-[600px]"
                leave-to-class="opacity-0 max-h-0"
              >
                <pre v-if="treeOpen" class="rounded-lg bg-slate-900 text-emerald-200 text-[12px] leading-relaxed font-mono p-3 overflow-x-auto whitespace-pre"><code>{{ sec.tree }}</code></pre>
              </Transition>
            </div>

            <!-- example: collapsible code block -->
            <div v-else-if="sec.kind === 'example'">
              <div class="rounded-lg overflow-hidden border border-slate-800 bg-slate-900">
                <div class="px-3.5 py-1.5 flex items-center gap-2 text-[10px] font-bold tracking-wider text-slate-400 uppercase border-b border-slate-800">
                  <div class="i-lucide-code-2 text-[12px]" />
                  Example
                </div>
                <CodeBlock :code="sec.code" lang="yaml" dark />
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script lang="ts">
function secKindIcon(sec: any): string {
  switch (sec.kind) {
    case 'anatomy': return 'i-lucide-list-tree'
    case 'loop':    return 'i-lucide-repeat'
    case 'flow':    return 'i-lucide-workflow'
    case 'tree':    return 'i-lucide-folder-tree'
    case 'example': return 'i-lucide-code-2'
    default:        return 'i-lucide-file-text'
  }
}
export { secKindIcon }
</script>
