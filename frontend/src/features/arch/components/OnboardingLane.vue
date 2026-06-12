<script setup lang="ts">
import ArchCard from './ArchCard.vue'
defineEmits<{ click: [key: string] }>()
</script>

<template>
  <div class="rounded-xl border border-emerald-200/60 bg-emerald-50/20 overflow-hidden">
    <div class="bg-emerald-500 h-1" />
    <div class="px-4 pt-3 pb-2"><span class="text-sm font-bold text-emerald-700">Onboarding</span></div>
    <div class="px-3 pb-2">
      <ArchCard icon="i-lucide-refresh-cw" title="Schema Sync" subtitle="SyncSchema() → rc_*" color="emerald" @click="$emit('click', 'onb-sync')" />
    </div>
    <div class="flex justify-center"><div class="i-lucide-arrow-down text-gray-300 text-xs" /></div>
    <div class="px-3 pb-2">
      <ArchCard icon="i-lucide-brain" title="Context Generator" subtitle="ReActExploreWorker" color="emerald" @click="$emit('click', 'onb-ctx')" />
    </div>
    <div class="flex justify-center"><div class="i-lucide-arrow-down text-gray-300 text-xs" /></div>
    <div class="px-3 pb-3">
      <ArchCard icon="i-lucide-vector" title="Embedding" subtitle="GenerateAndSaveEmbeddings()" color="emerald" @click="$emit('click', 'onb-emb')" />
    </div>
    <div class="flex justify-center pb-2">
      <div class="flex flex-col items-center text-xs text-emerald-400"><div class="i-lucide-arrow-down text-sm" />read/write<div class="i-lucide-arrow-up text-sm" /></div>
    </div>
  </div>
</template>
