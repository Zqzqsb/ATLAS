<script setup lang="ts">
import { computed, type Component } from 'vue'
import { ACCENTS } from '../../model/architecture'
import type { FlowDef } from '../../model/flows'
import OnboardingDetail from './modules/OnboardingDetail.vue'

const props = defineProps<{ flow: FlowDef }>()
const emit = defineEmits<{ back: [] }>()

const a = computed(() => ACCENTS[props.flow.accent])

// Registry: per-module internal architecture diagram, keyed by flow id.
const REGISTRY: Record<string, Component> = {
  onboarding: OnboardingDetail,
}
const detailComp = computed<Component | null>(() => REGISTRY[props.flow.id] ?? null)
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
        <h2 class="text-xl font-extrabold text-gray-900 m-0">{{ flow.title }}</h2>
        <p class="text-sm text-gray-500 mt-1 leading-snug">{{ flow.subtitle }}</p>
      </div>
    </div>

    <!-- Module internal architecture diagram -->
    <component :is="detailComp" v-if="detailComp" :flow="flow" />
    <div v-else class="text-center text-sm text-gray-400 py-16">
      <div class="i-lucide-construction text-2xl mx-auto mb-2" />
      该模块内部架构图建设中
    </div>
  </div>
</template>
