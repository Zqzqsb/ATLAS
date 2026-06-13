<script setup lang="ts">
import { ACCENTS, type AccentKey } from '../../../model/architecture'
import { computed } from 'vue'

const props = defineProps<{ icon: string; title: string; subtitle?: string; accent?: AccentKey }>()
const a = computed(() => ACCENTS[props.accent ?? 'slate'])
</script>

<template>
  <div class="flex items-center gap-2.5 mb-4">
    <div class="w-7 h-7 rounded-lg flex-center" :class="a.iconBg">
      <div :class="[icon, a.iconText, 'text-sm']" />
    </div>
    <div>
      <h3 class="text-base font-bold text-gray-900 leading-tight m-0">{{ title }}</h3>
      <p v-if="subtitle" class="text-xs text-gray-400 leading-snug">{{ subtitle }}</p>
    </div>
  </div>
</template>
