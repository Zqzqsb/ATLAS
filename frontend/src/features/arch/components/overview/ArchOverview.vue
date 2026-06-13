<script setup lang="ts">
import { ARCH_LAYERS, type ArchNode } from '../../model/architecture'
import ArchLayer from './ArchLayer.vue'

const emit = defineEmits<{ select: [node: ArchNode, ev: MouseEvent] }>()
</script>

<template>
  <div class="max-w-5xl mx-auto px-6 py-8">
    <!-- Title -->
    <div class="text-center mb-7">
      <h1 class="text-2xl font-extrabold text-gray-900 tracking-tight">ATLAS 全景架构</h1>
      <p class="text-sm text-gray-400 mt-1">点击带「展开」的模块，放大查看其内部 dataflow</p>
    </div>

    <!-- Layered stack with connectors -->
    <template v-for="(layer, idx) in ARCH_LAYERS" :key="layer.id">
      <ArchLayer :layer="layer" @select="(n, ev) => emit('select', n, ev)" />
      <!-- connector between layers -->
      <div v-if="idx < ARCH_LAYERS.length - 1" class="flex justify-center py-1.5">
        <div class="flex flex-col items-center text-gray-300">
          <div class="w-px h-3 bg-gray-300" />
          <div class="i-lucide-chevron-down text-sm -my-0.5" />
        </div>
      </div>
    </template>
  </div>
</template>
