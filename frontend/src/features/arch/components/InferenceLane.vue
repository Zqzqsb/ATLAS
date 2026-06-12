<script setup lang="ts">
import ArchCard from './ArchCard.vue'
defineEmits<{ click: [key: string] }>()
</script>

<template>
  <div class="rounded-xl border border-blue-200/60 bg-blue-50/20 overflow-hidden">
    <div class="bg-blue-500 h-1" />
    <div class="px-4 pt-3 pb-2"><span class="text-sm font-bold text-blue-700">Inference</span></div>
    <div class="px-3 pb-2">
      <ArchCard icon="i-lucide-git-graph" title="Adaptive Schema Linking" subtitle="Two-stage grounding pipeline" color="blue" @click="$emit('click', 'inf-link')" />
    </div>
    <div class="flex justify-center"><div class="i-lucide-arrow-down text-gray-300 text-xs" /></div>
    <div class="px-3 pb-2">
      <ArchCard icon="i-lucide-code" title="SQL Generator" subtitle="ReAct SQLGen Agent" color="blue" @click="$emit('click', 'inf-sql')" />
    </div>
    <div class="flex justify-center"><div class="i-lucide-arrow-down text-gray-300 text-xs" /></div>
    <div class="px-3 pb-3">
      <ArchCard icon="i-lucide-check-circle" title="Verify &amp; Execute" subtitle="DryRun → ExecuteQuery" color="blue" @click="$emit('click', 'inf-verify')" />
    </div>
    <div class="flex justify-center pb-2">
      <div class="flex flex-col items-center text-xs text-blue-400"><div class="i-lucide-arrow-down text-sm" />read/write<div class="i-lucide-arrow-up text-sm" /></div>
    </div>
  </div>
</template>
