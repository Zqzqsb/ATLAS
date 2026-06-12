<script setup lang="ts">
import ArchCard from './ArchCard.vue'
defineEmits<{ click: [key: string] }>()
</script>

<template>
  <div class="rounded-xl border border-amber-200/60 bg-amber-50/20 overflow-hidden">
    <div class="bg-amber-500 h-1" />
    <div class="px-4 pt-3 pb-2"><span class="text-sm font-bold text-amber-700">Self-Maintenance</span></div>
    <div class="px-3 pb-2">
      <ArchCard icon="i-lucide-radar" title="Change Detector" subtitle="DDLDetector" color="amber" @click="$emit('click', 'sm-detect')" />
    </div>
    <div class="flex justify-center"><div class="i-lucide-arrow-down text-gray-300 text-xs" /></div>
    <div class="px-3 pb-2">
      <ArchCard icon="i-lucide-hub" title="Coordinator" subtitle="AgentService" color="amber" @click="$emit('click', 'sm-coord')" />
    </div>
    <div class="flex justify-center"><div class="i-lucide-arrow-down text-gray-300 text-xs" /></div>
    <div class="px-3 pb-2">
      <ArchCard icon="i-lucide-wrench" title="Context Refresher" subtitle="ContextMaintainer" color="amber" @click="$emit('click', 'sm-refresh')" />
    </div>
    <div class="flex justify-center"><div class="i-lucide-arrow-down text-gray-300 text-xs" /></div>
    <div class="px-3 pb-3">
      <ArchCard icon="i-lucide-file-text" title="Change Logger" subtitle="ChangeLogger" color="amber" @click="$emit('click', 'sm-log')" />
    </div>
    <div class="flex justify-center pb-2">
      <div class="flex flex-col items-center text-xs text-amber-400"><div class="i-lucide-arrow-down text-sm" />read/write<div class="i-lucide-arrow-up text-sm" /></div>
    </div>
  </div>
</template>
