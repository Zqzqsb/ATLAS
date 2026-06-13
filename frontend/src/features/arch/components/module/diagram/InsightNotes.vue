<script setup lang="ts">
/**
 * Speaker-notes column cell: a stack of collapsible-style insight cards aligned
 * to one spine stage. Lives in the (hideable) left notes column — these read
 * like the hidden "presenter notes" behind a PPT slide.
 */
import { computed } from 'vue'
import { ACCENTS, type AccentKey } from '../../../model/architecture'
import type { Insight } from '../../../model/modules'

const props = defineProps<{
  accent: AccentKey
  /** plain intro sentence (e.g. the stage's one-liner) */
  intro?: string
  /** detailed insight cards */
  items?: Insight[]
}>()
const a = computed(() => ACCENTS[props.accent])
</script>

<template>
  <div class="space-y-2 border-l-2 border-dashed border-gray-200 pl-3">
    <p v-if="intro" class="text-[11px] text-gray-500 leading-relaxed">{{ intro }}</p>
    <div
      v-for="ins in items"
      :key="ins.title"
      class="rounded-xl border border-gray-200 bg-gradient-to-br from-white to-gray-50/60 px-3 py-2"
    >
      <div class="flex items-center gap-1.5 mb-0.5">
        <div class="w-5 h-5 rounded-md flex-center flex-shrink-0" :class="a.iconBg">
          <div :class="[ins.icon, a.iconText, 'text-[11px]']" />
        </div>
        <span class="text-xs font-bold text-gray-800">{{ ins.title }}</span>
      </div>
      <p class="text-[11px] text-gray-500 leading-relaxed">{{ ins.body }}</p>
    </div>
  </div>
</template>
