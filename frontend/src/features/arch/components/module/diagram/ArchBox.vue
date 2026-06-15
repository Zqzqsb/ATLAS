<script setup lang="ts">
import { computed } from 'vue'
import { ACCENTS, type AccentKey } from '../../../model/architecture'

const props = defineProps<{
  icon: string
  title: string
  /** short role label shown as a pill, e.g. "切分 + 分发" */
  role?: string
  accent: AccentKey
  /** badge on the top-right, e.g. "× N" */
  badge?: string
  /** muted = dashed/lighter (for input/boundary boxes) */
  muted?: boolean
}>()
const a = computed(() => ACCENTS[props.accent])
</script>

<template>
  <div
    class="rounded-2xl border bg-white overflow-hidden"
    :class="muted ? 'border-dashed border-gray-300' : `${a.surface}`"
  >
    <div class="h-1" :class="muted ? 'bg-gray-300' : a.bar" />
    <div class="px-4 py-3">
      <!-- header -->
      <div class="flex items-center gap-2.5 mb-2">
        <div class="w-8 h-8 rounded-lg flex-center flex-shrink-0" :class="a.iconBg">
          <div :class="[icon, a.iconText, 'text-base']" />
        </div>
        <div class="flex items-center gap-2 flex-1 min-w-0">
          <span class="text-sm font-bold text-gray-900">{{ title }}</span>
          <span v-if="role" class="px-2 py-0.5 rounded-full text-[11px] font-semibold border" :class="a.chip">{{ role }}</span>
        </div>
        <span v-if="badge" class="text-xs font-mono font-bold flex-shrink-0" :class="a.text">{{ badge }}</span>
        <slot name="refs" />
      </div>
      <slot />
    </div>
  </div>
</template>
