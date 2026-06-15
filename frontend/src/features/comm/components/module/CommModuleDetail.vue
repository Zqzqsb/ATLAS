<script setup lang="ts">
import { computed, ref } from 'vue'
import { ACCENTS } from '../../../arch/model/architecture'
import StageDetail from './StageDetail.vue'
import type { CommFlowDef } from '../../model/comm'

const props = defineProps<{ flow: CommFlowDef }>()
const emit = defineEmits<{ back: [] }>()

const a = computed(() => ACCENTS[props.flow.accent])
const showNotes = ref(true)
</script>

<template>
  <div class="mx-auto px-6 py-7 transition-[max-width] duration-300" :class="showNotes ? 'max-w-7xl' : 'max-w-5xl'">
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

    <StageDetail :flow="flow" :show-notes="showNotes" />
  </div>
</template>
