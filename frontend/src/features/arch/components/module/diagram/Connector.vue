<script setup lang="ts">
/** Vertical connector between architecture boxes: line + arrow + optional label,
 *  with a subtle downward "data flow" dot. */
defineProps<{ label?: string }>()
</script>

<template>
  <div class="flex justify-center py-1.5">
    <div class="relative flex flex-col items-center">
      <!-- track -->
      <div class="relative w-px h-7 bg-gray-200 overflow-hidden">
        <div class="flow-dot absolute left-1/2 -translate-x-1/2 w-1 h-1 rounded-full bg-gray-400" />
      </div>
      <!-- label -->
      <div
        v-if="label"
        class="absolute left-3 top-1/2 -translate-y-1/2 whitespace-nowrap text-[11px] font-mono text-gray-400 px-1.5 py-0.5 rounded bg-gray-50 border border-gray-100"
      >
        {{ label }}
      </div>
      <div class="i-lucide-chevron-down text-gray-300 text-sm -mt-1" />
    </div>
  </div>
</template>

<style scoped>
.flow-dot {
  animation: flow-down 2.4s ease-in-out infinite;
}
@keyframes flow-down {
  0% { top: -4px; opacity: 0; }
  25% { opacity: 1; }
  75% { opacity: 1; }
  100% { top: 28px; opacity: 0; }
}
</style>
