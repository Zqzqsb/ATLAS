<script setup lang="ts">
import { computed, type Component } from 'vue'
import { ACCENTS } from '../../model/architecture'
import type { FlowDef } from '../../model/flows'
import DataflowStepper from './DataflowStepper.vue'
import OnboardingDetail from './modules/OnboardingDetail.vue'

const props = defineProps<{ flow: FlowDef }>()
const emit = defineEmits<{ back: [] }>()

const a = computed(() => ACCENTS[props.flow.accent])

// Registry: per-module detail composition. Falls back to the bare dataflow
// stepper for modules that only have flow data so far.
const REGISTRY: Record<string, Component> = {
  onboarding: OnboardingDetail,
}
const detailComp = computed<Component>(() => REGISTRY[props.flow.id] ?? DataflowStepper)
</script>

<template>
  <div class="max-w-5xl mx-auto px-6 py-7">
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
        <div class="flex items-center gap-2">
          <h2 class="text-xl font-extrabold text-gray-900 m-0">{{ flow.title }}</h2>
          <span class="px-2 py-0.5 rounded-full text-xs font-semibold border" :class="a.chip">{{ flow.steps.length }} 步</span>
        </div>
        <p class="text-sm text-gray-500 mt-1 leading-snug">{{ flow.subtitle }}</p>
      </div>
    </div>

    <!-- Pipeline ribbon (horizontal flow at a glance) -->
    <div class="flex items-center gap-1 mb-7 overflow-x-auto pb-1">
      <template v-for="(step, idx) in flow.steps" :key="step.id">
        <div class="flex items-center gap-2 px-3 py-1.5 rounded-lg border bg-white flex-shrink-0" :class="ACCENTS[step.accent].surface">
          <div :class="[step.icon, ACCENTS[step.accent].text, 'text-sm']" />
          <span class="text-xs font-semibold text-gray-700 whitespace-nowrap">{{ step.title }}</span>
        </div>
        <div v-if="idx < flow.steps.length - 1" class="i-lucide-arrow-right text-gray-300 flex-shrink-0" />
      </template>
    </div>

    <!-- Module-specific detailed architecture -->
    <component :is="detailComp" :flow="flow" />
  </div>
</template>
