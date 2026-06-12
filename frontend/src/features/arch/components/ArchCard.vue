<script setup lang="ts">
defineProps<{ icon: string; title: string; subtitle: string; color: string }>()
defineEmits<{ click: [] }>()
</script>

<template>
  <button
    class="w-full rounded-lg bg-white border border-gray-200 shadow-sm p-3 text-left transition-all hover:shadow-md hover:border-gray-300 cursor-pointer"
    @click="$emit('click')"
  >
    <div class="flex items-center gap-2">
      <div class="w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0" :class="`bg-${color}-100`">
        <div :class="[icon, `text-${color}-600 text-sm`]" />
      </div>
      <div class="flex-1 min-w-0">
        <div class="text-xs font-semibold text-gray-800">{{ title }}</div>
        <div class="text-xs text-gray-400">{{ subtitle }}</div>
      </div>
      <div class="i-lucide-arrow-up-right text-gray-300 text-xs flex-shrink-0" />
    </div>
  </button>
</template>
