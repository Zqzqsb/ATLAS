<script setup lang="ts">
defineProps<{ label: string; tag: string; desc: string }>()
</script>

<template>
  <div>
    <h3 class="text-base font-bold text-gray-900 mb-1">{{ label }}</h3>
    <div class="text-xs text-gray-400 font-mono mb-3">{{ tag }}</div>
    <p class="text-sm text-gray-600">{{ desc }}</p>
  </div>
</template>
