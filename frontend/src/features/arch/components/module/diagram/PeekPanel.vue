<script setup lang="ts">
import { ref, computed } from 'vue'
import { ACCENTS, type AccentKey } from '../../../model/architecture'

const props = defineProps<{
  label: string
  icon?: string
  count?: number
  accent?: AccentKey
  /** start expanded */
  open?: boolean
}>()
const expanded = ref(props.open ?? false)
const a = computed(() => ACCENTS[props.accent ?? 'slate'])
</script>

<template>
  <div class="rounded-lg border bg-white/70" :class="expanded ? a.surface : 'border-gray-200'">
    <button
      class="w-full flex items-center gap-2 px-2.5 py-1.5 text-left"
      @click="expanded = !expanded"
    >
      <div v-if="icon" :class="[icon, a.text, 'text-xs flex-shrink-0']" />
      <span class="text-xs font-semibold text-gray-700">{{ label }}</span>
      <span v-if="count != null" class="px-1.5 rounded-full text-[10px] font-bold" :class="a.chip">{{ count }}</span>
      <div
        class="i-lucide-chevron-down text-gray-400 text-xs ml-auto transition-transform"
        :class="{ 'rotate-180': expanded }"
      />
    </button>
    <Transition name="peek">
      <div v-if="expanded" class="px-2.5 pb-2.5 pt-0.5">
        <slot />
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.peek-enter-active,
.peek-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.peek-enter-from,
.peek-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
