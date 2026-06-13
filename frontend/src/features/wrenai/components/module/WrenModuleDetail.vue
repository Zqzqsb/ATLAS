<script setup lang="ts">
import { computed, ref, type Component } from 'vue'
import { ACCENTS } from '../../../arch/model/architecture'
import type { WrenFlowDef } from '../../model/wren'
import MdlDetail from './modules/MdlDetail.vue'
import QueryFlowDetail from './modules/QueryFlowDetail.vue'
import PlanningDetail from './modules/PlanningDetail.vue'
import MemoryDetail from './modules/MemoryDetail.vue'
import SkillsDetail from './modules/SkillsDetail.vue'
import ExecutionDetail from './modules/ExecutionDetail.vue'

const props = defineProps<{ flow: WrenFlowDef }>()
const emit = defineEmits<{ back: [] }>()

const a = computed(() => ACCENTS[props.flow.accent])

// "Presenter notes" column — on by default; toggle to hide like PPT speaker notes.
const showNotes = ref(true)

const REGISTRY: Record<string, Component> = {
  mdl: MdlDetail,
  query: QueryFlowDetail,
  planning: PlanningDetail,
  memory: MemoryDetail,
  skills: SkillsDetail,
  execution: ExecutionDetail,
}
const detailComp = computed<Component | null>(() => REGISTRY[props.flow.id] ?? null)
</script>

<template>
  <div class="mx-auto px-6 py-7 transition-[max-width] duration-300" :class="showNotes ? 'max-w-7xl' : 'max-w-5xl'">
    <!-- Header -->
    <div class="flex items-start gap-3 mb-6">
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

    <component :is="detailComp" v-if="detailComp" :flow="flow" :show-notes="showNotes" />
    <div v-else class="text-center text-sm text-gray-400 py-16">
      <div class="i-lucide-construction text-2xl mx-auto mb-2" />
      该模块内部架构图建设中
    </div>
  </div>
</template>
