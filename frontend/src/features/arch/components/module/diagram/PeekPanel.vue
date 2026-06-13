<script setup lang="ts">
import { ref, computed } from 'vue'
import { NPopover } from 'naive-ui'
import { ACCENTS, type AccentKey } from '../../../model/architecture'

const props = defineProps<{
  label: string
  icon?: string
  count?: number
  accent?: AccentKey
}>()
const open = ref(false)
const a = computed(() => ACCENTS[props.accent ?? 'slate'])
</script>

<template>
  <NPopover
    v-model:show="open"
    trigger="click"
    placement="bottom-start"
    :width="320"
    :content-style="{ padding: '0' }"
  >
    <template #trigger>
      <button
        type="button"
        class="w-full flex items-center gap-2 px-2.5 py-1.5 text-left rounded-lg border bg-white/70 transition-colors"
        :class="open ? a.surface : 'border-gray-200 hover:border-gray-300'"
      >
        <div v-if="icon" :class="[icon, a.text, 'text-xs flex-shrink-0']" />
        <span class="text-xs font-semibold text-gray-700">{{ label }}</span>
        <span v-if="count != null" class="px-1.5 rounded-full text-[10px] font-bold" :class="a.chip">{{ count }}</span>
        <div
          class="i-lucide-chevron-down text-gray-400 text-xs ml-auto transition-transform"
          :class="{ 'rotate-180': open }"
        />
      </button>
    </template>

    <!-- floating panel (teleported to body — no layout reflow) -->
    <div class="rounded-lg overflow-hidden">
      <div class="flex items-center gap-2 px-3 py-2 border-b border-gray-100" :class="a.surface">
        <div v-if="icon" :class="[icon, a.text, 'text-sm flex-shrink-0']" />
        <span class="text-xs font-bold text-gray-800">{{ label }}</span>
        <span v-if="count != null" class="px-1.5 rounded-full text-[10px] font-bold ml-auto" :class="a.chip">{{ count }}</span>
      </div>
      <div class="px-3 py-2.5 max-h-[60vh] overflow-auto">
        <slot />
      </div>
    </div>
  </NPopover>
</template>
